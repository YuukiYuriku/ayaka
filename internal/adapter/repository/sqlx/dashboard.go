package sqlx

import (
	"context"
	"fmt"


	"gitlab.com/ayaka/internal/adapter/repository"
	"gitlab.com/ayaka/internal/domain/dashboard"
)

type DashboardRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *DashboardRepository) Fetch(ctx context.Context) (*dashboard.Read, error) {
	var dashboardRead dashboard.Read

	var outstandingMaterialTransfer uint
	query := `SELECT COUNT(*) FROM (
		SELECT 
			t.DocNo AS DocNoMaterialTransfer,
			h.DocDt,
			t.ItCode,
			i.ItName,
			t.BatchNo,
			(t.Qty - COALESCE(SUM(r.QtyActual), 0)) AS QtyRemaining,
			u.UomName
		FROM tblmaterialtransferdtl t
		JOIN tblmaterialtransferhdr h ON h.DocNo = t.DocNo
		JOIN tblitem i ON i.ItCode = t.ItCode
		JOIN tblUom u on i.PurchaseUomCode = u.UomCode
		LEFT JOIN tblmaterialreceivedtl r 
			ON r.DocNoMaterialTransfer = t.DocNo
			AND r.ItCode = t.ItCode
			AND (r.BatchNo = t.BatchNo OR (r.BatchNo IS NULL AND t.BatchNo IS NULL)) -- handle batch null
		WHERE t.CancelInd = 'N'
		GROUP BY 
				t.DocNo,
				t.DNo,
				t.ItCode,
				i.ItName,
				t.BatchNo,
				t.Qty,
				h.WhsCodeFrom,
				h.WhsCodeTo
		HAVING QtyRemaining > 0
	) as grouped`

	if err := t.DB.GetContext(ctx, &outstandingMaterialTransfer, query); err != nil {
		return nil, fmt.Errorf("error counting records: %w", err)
	}


	var outstandingPurchaseOrder uint
	query = `SELECT COUNT(*) FROM (
		SELECT 
			d.DocNo,
			h.Status,
			v.VendorName,
			mrh.Department,
			i.ItName,
			d.Qty AS PurchaseQty,
			d.Qty - COALESCE((
				SELECT SUM(r.PurchaseQty)
				FROM tblpurchasematerialreceivedtl r
				WHERE r.PurchaseOrderDocNo = d.DocNo
				AND r.PurchaseOrderDNo = d.DNo
				AND (r.CancelInd IS NULL OR r.CancelInd != 'Y')
			), 0) AS OutstandingQty,
			u.UomName,
			c.CurName,
			(d.Total / d.Qty) AS Price,
			d.Total
		FROM 
			tblpurchaseorderdtl d
		JOIN 
			tblpurchaseorderhdr h ON d.DocNo = h.DocNo
        JOIN tblpurchaseorderreqdtl por
        	ON d.PurchaseOrderReqDocNo = por.DocNo
            AND d.PurchaseOrderReqDNo = por.DNo
        JOIN tblmaterialrequestdtl mr
        	ON por.MaterialReqDocNo = mr.DocNo
            AND por.MaterialReqDNo = mr.DNo
        JOIN tblmaterialrequesthdr mrh
        	ON mr.DocNo = mrh.DocNo
        JOIN tblvendorquotationhdr vq
        	ON por.VendorQTDocNo = vq.DocNo
		JOIN 
			tblitem i ON d.ItCode = i.ItCode
		JOIN 
			tbluom u ON i.PurchaseUomCode = u.UomCode
		JOIN tblvendorhdr v
			ON h.VendorCode = v.VendorCode
		JOIN tblcurrency c
			ON vq.CurCode = c.CurCode
		WHERE 
			(d.CancelInd != 'Y')
		HAVING 
			OutstandingQty > 0
	) AS grouped`

	if err := t.DB.GetContext(ctx, &outstandingPurchaseOrder, query); err != nil {
		return nil, fmt.Errorf("error counting records: %w", err)
	}

	var outstandingMaterialOrder uint
	query = `SELECT COUNT(*) FROM (
		SELECT 
				mr.DocNo,
				i.ItName,
				(mr.Qty - COALESCE(SUM(r.PurchaseQty), 0)) AS OutstandingQty,
				mr.Qty RequestedQty
			FROM tblmaterialrequestdtl mr
			JOIN tblitem i ON mr.ItCode = i.ItCode
			JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode
			JOIN tblmaterialrequesthdr h ON mr.DocNo = h.DocNo
			LEFT JOIN tblpurchaseorderreqdtl por
				ON mr.DocNo = por.MaterialReqDocNo AND mr.DNo = por.MaterialReqDNo
			LEFT JOIN tblpurchaseorderdtl po
				ON por.DocNo = po.PurchaseOrderReqDocNo AND por.DNo = po.PurchaseOrderReqDNo
			LEFT JOIN tblpurchasematerialreceivedtl r
				ON po.DocNo = r.PurchaseOrderDocNo AND po.DNo = r.PurchaseOrderDNo
			WHERE mr.CancelInd != 'Y' AND mr.OpenInd = 'Y'
			GROUP BY mr.DocNo, mr.DNo, mr.ItCode, mr.Qty
			HAVING OutstandingQty > 0
		) AS grouped`

	if err := t.DB.GetContext(ctx, &outstandingMaterialOrder, query); err != nil {
		return nil, fmt.Errorf("error counting records: %w", err)
	}

	dashboardRead.OutstandingMaterialOrder = outstandingMaterialOrder
	dashboardRead.OutstandingMaterialTransfer = outstandingMaterialTransfer
	dashboardRead.OutstandingPurchaseOrder = outstandingPurchaseOrder

	return &dashboardRead, nil
}