package tbluser

type Logintbluser struct {
	CompanyID string `db:"CompanyID" json:"company_id" validate:"required" label:"Company ID"`
	UserCode  string `db:"UserCode" json:"user_code" validate:"required" label:"User Code"`
	UserName  string `db:"UserName" json:"user_name" validate:"required" label:"User Name"`
	Password  string `db:"Pwd" json:"password" validate:"required" label:"Password"`
}

type ForgotPassword struct {
	Password string `db:"Pwd" json:"new_password" validate:"complexpassword" label:"New Password"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password" label:"Confirm Password"`
}