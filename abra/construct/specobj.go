package construct

import (
	"fmt"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/lunfardo314/goq/abra"
	"github.com/lunfardo314/goq/abra/validate"
)

// finds or creates site which returns constant 'val'
// creates respective LUT if necessary

func Get1TritConstSite(codeUnit *CodeUnit, branch *Branch, val int8) *Site {
	// first try to find if there's constant lut site for val
	// if not, create one
	lookupName := fmt.Sprintf("1trit_const_site_%s", TritName(val))
	ret := FindSite(branch, lookupName)
	if ret != nil {
		return ret
	}
	// didn't find. Need to create one
	// first find or create the only lut for 1 trit constant
	lutRepr := Get1TritConstLutRepr(val)
	lutValConstBlock := GetLUTBlock(codeUnit, lutRepr)

	// now create site in the branch
	// it will always generate constant trit
	// the input for lut is 3 repeated 1-trit sites from lst of the branche's input
	any := GetAnyTritSite(codeUnit, branch)
	ret = NewKnotSite(1, lookupName, lutValConstBlock, any, any, any)
	MustAddNewSite(branch, ret)
	return ret
}

// returns site which always produces first trit from first input in the branch

func GetAnyTritSite(codeUnit *CodeUnit, branch *Branch) *Site {
	lookupName := "any_trit_site" // each branch will have the only site with this name
	ret := FindSite(branch, lookupName)
	if ret != nil {
		return ret
	}
	// must be the only LstSliceBlock in the code unit
	lstBlock := GetSliceBranchBlock(codeUnit, GetInputSite(branch, 0).Size, 0, 1)
	ret = NewKnotSite(1, lookupName, lstBlock, GetInputSite(branch, 0))
	MustAddNewSite(branch, ret)
	return ret
}

// returns site which always produces specific trit vector

func GetTritVectorConstSite(codeUnit *CodeUnit, branch *Branch, val Trits) *Site {
	lookupName := TritsToString(val) + "_const_site"
	ret := FindSite(branch, lookupName)
	if ret != nil {
		return ret
	}
	inputs := make([]*Site, len(val))
	for i, trit := range val {
		inputs[i] = Get1TritConstSite(codeUnit, branch, trit)
	}

	concatBlock := GetConcatBlockForSize(codeUnit, len(val))
	ret = NewKnotSite(len(val), lookupName, concatBlock, inputs...)
	MustAddNewSite(branch, ret)
	return ret
}

// find or create branch which concatenates its inputs of the specified total size

func GetConcatBlockForSize(codeUnit *CodeUnit, size int) *Block {
	lookupName := fmt.Sprintf("concat_branch_%d", size)
	ret := FindBranchBlock(codeUnit, lookupName)
	if ret != nil {
		return ret
	}
	ret = MustAddNewBranchBlock(codeUnit, lookupName, size)
	input := AddInputSite(ret.Branch, size)
	output := NewMergeSite(size, "", input)

	output.SiteType = SITE_OUTPUT
	MustAddNewSite(ret.Branch, output)

	validate.AssertValid(ret.Branch)
	return ret
}

// returns or creates block which takes to output least significant trit of it input
// due to requirement to have exact size matches, there's one block per each
// suboptimal !!!

func GetSliceBranchBlock(codeUnit *CodeUnit, inputSize, offset, size int) *Block {
	if size == 0 {
		panic("GetSliceBranchBlock: zero sized slice not allowed")
	}
	if offset+size > inputSize {
		panic("GetSliceBranchBlock: wrong arguments")
	}
	lookupName := fmt.Sprintf("slice_branch_%d_%d_%d", inputSize, offset, size)
	ret := FindBranchBlock(codeUnit, lookupName)
	if ret != nil {
		return ret
	}
	ret = MustAddNewBranchBlock(codeUnit, lookupName, size)

	if offset != 0 {
		AddInputSite(ret.Branch, offset)
	}
	theSlice := AddInputSite(ret.Branch, size)
	if offset+size < inputSize {
		AddInputSite(ret.Branch, inputSize-offset-size)
	}
	output := NewMergeSite(size, "", theSlice)
	output.SiteType = SITE_OUTPUT
	MustAddNewSite(ret.Branch, output)

	validate.AssertValid(ret.Branch)
	return ret
}

// get nullify LUT block for true or false
func GetNullifyLUTBlock(codeUnit *CodeUnit, trueFalse bool) *Block {
	strRepr := GetNullifyLUTRepr(trueFalse)
	return GetLUTBlock(codeUnit, strRepr)
}

// finds of creates nullify branch block

func GetNullifyBranchBlock(codeUnit *CodeUnit, size int, trueFalse bool) *Block {
	lookupName := fmt.Sprintf("nullify_%v_arg_%d", trueFalse, size)
	ret := FindBranchBlock(codeUnit, lookupName)
	if ret != nil {
		return ret
	}
	ret = MustAddNewBranchBlock(codeUnit, lookupName, size)
	AddInputSite(ret.Branch, 1) // condition
	for i := 0; i < size; i++ {
		AddInputSite(ret.Branch, 1) // arg
	}
	nullifyLutBlock := GetNullifyLUTBlock(codeUnit, trueFalse)
	condInput := GetInputSite(ret.Branch, 0)

	for i := 0; i < size; i++ {
		nullifyTritSite :=
			NewKnotSite(1, "", nullifyLutBlock, condInput, GetInputSite(ret.Branch, i+1), condInput)
		ChangeSiteType(nullifyTritSite, SITE_OUTPUT)
		MustAddNewSite(ret.Branch, nullifyTritSite)
	}
	validate.AssertValid(ret.Branch)
	return ret
}
