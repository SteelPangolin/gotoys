package arclight

import "testing"

func strSlicesEqual(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

// Case where no ancestors exist in the input
func TestImplicitDirs_NoAncestor(t *testing.T) {
	input := []string{
		"alpha/beta/delta",
		"alpha/beta/gamma",
		"alpha/mu/nu",
	}
	output := ImplicitDirs(input)
	expected := []string{
		"alpha",
		"alpha/beta",
		"alpha/mu",
	}
	if !strSlicesEqual(output, expected) {
		t.Errorf("output %#v != expected %#v", output, expected)
	}
}

// Case where an ancestor exists in the input
func TestImplicitDirs_Ancestor(t *testing.T) {
    input := []string{
        "theta",
        "theta/phi/psi/upsilon",
    }
    output := ImplicitDirs(input)
    expected := []string{
        "theta/phi",
        "theta/phi/psi",
    }
    if !strSlicesEqual(output, expected) {
        t.Errorf("output %#v != expected %#v", output, expected)
    }
}
