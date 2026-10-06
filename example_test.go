package houjin_test

import (
	"fmt"
	"os"

	houjin "github.com/sibukixxx/company-catalog"
)

func ExampleReadCompanySourcePackage() {
	f, err := os.Open("docs/examples/company-source-package.v1.json")
	if err != nil {
		panic(err)
	}
	p, err := houjin.ReadCompanySourcePackage(f)
	closeErr := f.Close()
	if err != nil {
		panic(err)
	}
	if closeErr != nil {
		panic(closeErr)
	}
	fmt.Println(p.SchemaVersion)
	fmt.Println(p.Manifest.RecordCount, p.Completeness)
	// Output:
	// techvit.company.source-package/v1
	// 2 COMPLETE
}
