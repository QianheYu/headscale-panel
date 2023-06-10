package common

import (
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ch_translations "github.com/go-playground/validator/v10/translations/zh"
	"headscale-panel/log"
	"regexp"
)

// Global Validate data validation real column
var Validate *validator.Validate

// Global Translator
var Trans ut.Translator

// Initialization of Validator data validation
func InitValidate() {
	chinese := zh.New()
	uni := ut.New(chinese, chinese)
	trans, _ := uni.GetTranslator("zh")
	Trans = trans
	Validate = validator.New()
	_ = ch_translations.RegisterDefaultTranslations(Validate, Trans)
	_ = Validate.RegisterValidation("checkMobile", checkMobile)
	log.Log.Infof("Initialisation of validator.v10 data verifier complete")
}

func checkMobile(fl validator.FieldLevel) bool {
	reg := `^1([38][0-9]|14[579]|5[^4]|16[6]|7[1-35-8]|9[189])\d{8}$`
	rgx := regexp.MustCompile(reg)
	return rgx.MatchString(fl.Field().String())
}

func checkEmail(fl validator.FieldLevel) bool {
	reg := `^(([a-zA-Z]|[0-9])+\.)?([a-zA-Z]|[0-9])+@[a-zA-Z0-9]+\.([a-zA-Z]{2,4})$`
	rgx := regexp.MustCompile(reg)
	return rgx.MatchString(fl.Field().String())
}
