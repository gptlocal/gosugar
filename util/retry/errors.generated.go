// Code generated by codegen.errors. DO NOT EDIT.
package retry
import "github.com/gptlocal/gosugar/errors"

type errPathObjHolder struct{}

func newError(values ...interface{}) *errors.Error {
	return errors.New(values...).WithPathObj(errPathObjHolder{})
}
