package main

import (
	"math"
	"testing"
)

// TODO maybe redact my path

var successOutput = `Gobra 1.1-SNAPSHOT (46a35ffb@(detached))
(c) Copyright ETH Zurich 2012 - 2024
14:21:12.006 [main] INFO viper.gobra.Gobra - Verifying package /home/ali/Code/gobra-book-monorepo/scratch - bank [14:21:12]
14:21:15.583 [thread-5] INFO viper.gobra.Gobra - Gobra found no errors
14:21:15.583 [main] INFO viper.gobra.Gobra - Gobra has found 0 error(s)
14:21:15.604 [Thread-0] INFO viper.gobra.Gobra - Writing report to .gobra/stats.json`

var failedOutput = `Gobra 1.1-SNAPSHOT (46a35ffb@(detached))
(c) Copyright ETH Zurich 2012 - 2024
14:50:32.055 [main] INFO viper.gobra.Gobra - Verifying package /home/ali/Code/gobra-book-monorepo/scratch - main [14:50:32]
14:50:35.387 [ForkJoinPool-3-worker-2] ERROR viper.gobra.reporting.FileWriterReporter - Error at: <abs.go:29:8> Precondition of call Abs(MinInt) might not hold.
Assertion x != MinInt might not hold.
14:50:35.403 [thread-2] ERROR viper.gobra.Gobra - Gobra has found 1 error(s) in package /home/ali/Code/gobra-book-monorepo/scratch - main
14:50:35.403 [main] INFO viper.gobra.Gobra - Gobra has found 1 error(s)
14:50:35.427 [Thread-0] INFO viper.gobra.Gobra - Writing report to .gobra/stats.json
java -jar -Xss1g /home/ali/Code/gobra.jar -i abs.go`

func TestVerified(t *testing.T) {
	resp, err := ParseGobraOutput(successOutput)

	if err != nil {
		t.Fatal(err)
	}

	if !resp.Verified {
		t.Fatal("Failed to parse that verification succeeded")
	}

	if resp.MainError.ErrorMessage != "" {
		t.Fatal("There is an error message when there should be none")
	}
}

func TestFailDetected(t *testing.T) {
	resp, err := ParseGobraOutput(failedOutput)

	if err != nil {
		t.Fatal(err)
	}

	if resp.Verified {
		t.Fatal("Failed to parse that verification failed")
	}

	if resp.MainError.ErrorMessage != "Assertion x != MinInt might not hold." {
		t.Fatalf("Failed to parse error message, got %s", resp.MainError.ErrorMessage)
	}
}
