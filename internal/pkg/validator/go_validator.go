package validator

import (
	"context"
	"unicode"

	// "errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"gitlab.com/ayaka/internal/adapter/repository"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type GoValidator struct {
	validate *validator.Validate
	uni      ut.Translator
	DB       *repository.Sqlx `inject:"database"`
}

type ValidationError struct {
	ErrorFields map[string]string `json:"error_fields,omitempty"`
}

func NewGoValidator() *GoValidator {
	v := validator.New()
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")

	en_translations.RegisterDefaultTranslations(v, trans)

	// Override default required message
	v.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} is a required field", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		field := toProperCase(fe.Field())
		return fmt.Sprintf("%s is a required field", field)
	})

	// Override default min message
	v.RegisterTranslation("min", trans, func(ut ut.Translator) error {
		return ut.Add("min", "{0} must be at least {1} characters long", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		field := toProperCase(fe.Field())
		return fmt.Sprintf("%s must be at least %s characters long", field, fe.Param())
	})

	// Override default max message
	v.RegisterTranslation("max", trans, func(ut ut.Translator) error {
		return ut.Add("max", "{0} must be at most {1} characters long", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		field := toProperCase(fe.Field())
		return fmt.Sprintf("%s must be at most %s characters long", field, fe.Param())
	})

	v.RegisterTranslation("eqfield", trans, func(ut ut.Translator) error {
		return ut.Add("eqfield", "{0} must be equal to {1}", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		field := toProperCase(fe.Field())
		other := toProperCase(fe.Param())
		return fmt.Sprintf("%s must be equal to %s", field, other)
	})

	v.RegisterTranslation("email", trans, func(ut ut.Translator) error {
		return ut.Add("email", "{0} must be a valid email address", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		field := toProperCase(fe.Field())
		return fmt.Sprintf("%s must be a valid email address", field)
	})

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	v.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if nullData, ok := field.Interface().(nulldatatype.NullDataType); ok {
			if !nullData.Valid {
				return ""
			}
			return nullData.String
		}
		return nil
	}, nulldatatype.NullDataType{})

	gv := &GoValidator{validate: v, uni: trans}
	gv.registerCustomValidators()

	return gv
}

func (v *GoValidator) registerCustomValidators() {
	// Register the validation functions
	v.validate.RegisterValidation("unique", v.uniqueValidator)
	v.validate.RegisterValidation("complexpassword", v.complexPasswordValidator)
	v.validate.RegisterValidation("incolumn", v.incolumnValidator)

	// unique validate
	v.validate.RegisterTranslation("unique", v.uni, func(ut ut.Translator) error {
		return ut.Add("unique", "{0} already exists", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		return fmt.Sprintf("%s already exists", fe.Value().(string))
	})

	// incolumn validate
	v.validate.RegisterTranslation("incolumn", v.uni, func(ut ut.Translator) error {
		return ut.Add("incolumn", "{0} not exists", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		return fmt.Sprintf("%s not exists", fe.Value().(string))
	})

	//complex pass
	v.validate.RegisterTranslation("complexpassword", v.uni, func(ut ut.Translator) error {
		return ut.Add("complexpassword", "{0} not valid", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		return "Password must be 8-12 characters long, contain: uppercase, lowercase, special character and number"
	})
}

func (v *GoValidator) uniqueValidator(fl validator.FieldLevel) bool {
	params := strings.Split(fl.Param(), "->")
	if len(params) != 2 {
		return false
	}

	tableName := params[0]
	fieldName := params[1]
	fieldValue := fl.Field().String()

	var count int
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s = ?", tableName, fieldName)
	err := v.DB.Get(&count, query, fieldValue)

	if err != nil {
		return false
	}

	return count == 0
}

func (v *GoValidator) incolumnValidator(fl validator.FieldLevel) bool {
	params := strings.Split(fl.Param(), "->")
	if len(params) != 2 {
		return false
	}

	tableName := params[0]
	fieldName := params[1]

	// Handle NullDataType
	var fieldValue string

	// Check if the field is NullDataType
	if nullData, ok := fl.Field().Interface().(nulldatatype.NullDataType); ok {
		// If it's null or empty, skip validation
		if !nullData.Valid || nullData.String == "" {
			return true
		}
		fieldValue = nullData.String
	} else {
		// For regular string fields
		fieldValue = fl.Field().String()
	}

	// Skip validation if field is empty
	if fieldValue == "" {
		return true
	}

	var count int
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s = ?", tableName, fieldName)
	err := v.DB.Get(&count, query, fieldValue)

	if err != nil {
		return false
	}

	return count != 0
}

func (v *GoValidator) complexPasswordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 || len(password) > 12 {
		return false
	}

	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// Check for at least one number
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_\-+={[}\]|\\:;"'<,>.?/~]`).MatchString(password)

	// Check that all conditions are met
	return hasLower && hasUpper && hasNumber && hasSpecial
}

func (v *GoValidator) Validate(ctx context.Context, data interface{}) error {
	err := v.validate.StructCtx(ctx, data)
	if err == nil {
		return nil
	}

	if _, ok := err.(*validator.InvalidValidationError); ok {
		return err
	}

	validationErrors := err.(validator.ValidationErrors)
	if len(validationErrors) > 0 {
		errorFields := make(map[string]string)

		for _, err := range validationErrors {
			// Using the JSON tag name directly from the validator
			fieldName := err.Field()
			errorFields[fieldName] = err.Translate(v.uni)
		}

		return &ValidationError{
			ErrorFields: errorFields,
		}
	}

	return nil
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	var errMsgs []string
	for field, msg := range ve.ErrorFields {
		errMsgs = append(errMsgs, fmt.Sprintf("%s: %s", field, msg))
	}
	return strings.Join(errMsgs, "; ")
}

func toProperCase(input string) string {
	// Split the input string by underscore
	words := strings.Split(input, "_")

	// Capitalize first letter of each word
	for i, word := range words {
		if len(word) > 0 {
			// Convert word to rune slice for proper unicode handling
			runes := []rune(word)
			// Capitalize first letter
			runes[0] = unicode.ToUpper(runes[0])
			// Convert remaining letters to lower case
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
			words[i] = string(runes)
		}
	}

	// Join words with space
	return strings.Join(words, " ")
}
