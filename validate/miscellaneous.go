package validate

import (
	"fmt"
	"regexp"
)

type StringLengthValidator struct {
	Value       string
	LessThan    int
	GreaterThan int
	FieldName   string
}

func (v *StringLengthValidator) Check() (string,bool) {
	if len(v.Value) > v.LessThan {
		return fmt.Sprintf("[%s] length must less than %d",v.FieldName,v.LessThan),false
	}
	if len(v.Value) < v.GreaterThan {
		return fmt.Sprintf("[%s] length must greater than %d",v.FieldName,v.GreaterThan),false
	}
	return "",true
}

type EmailValidator struct {
	Value string
	FieldName string
}

func (v *EmailValidator) Check() (string, bool) {
	if len(v.FieldName) == 0{
		v.FieldName =  "email"
	}
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	isValidate := re.Match([]byte(v.Value))
	if isValidate {
		return "",true
	}else{
		return fmt.Sprintf("[%s] is not validate email address",v.FieldName),false
	}
}
