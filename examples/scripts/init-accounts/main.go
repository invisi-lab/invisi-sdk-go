// Create fresh local demonstration accounts; never overwrite an existing file.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
)

func main() {
	path := flag.String("out", shared.Settings.AccountsFile, "new local account file")
	flag.Parse()
	if err := shared.InitializeAccounts(*path); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created ten fresh local accounts (Alice through Judy):", *path)
	fmt.Printf("Prepare approximately %s %s for each submitter %v, plus the token funding required by your chosen examples.\n", shared.Settings.SuggestedSubmitterGasBalance, shared.Settings.NativeCurrencySymbol, shared.Settings.SubmitterNames)
	fmt.Println("Only recipients need no native currency. Initialization never funds accounts or sends transactions.")
}
