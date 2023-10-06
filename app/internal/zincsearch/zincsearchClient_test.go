package zincsearch

import (
	"testing" // Add the import path for the testing package
)


func Test_checkClient(t *testing.T) {
	r := NewZincSearchClient()
	err := r.CheckClient()

	if err != nil {
		t.Errorf("la validación falló con el error: %v", err)
		t.Fail()
	}
}
