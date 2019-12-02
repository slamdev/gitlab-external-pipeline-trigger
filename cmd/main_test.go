package main

import (
	"flag"
	"os"
	"testing"
)

//noinspection GoUnhandledErrorResult
func Test_Should_Run(t *testing.T) {
	t.Skip("skip integration test with real data")
	flag.Set("u-token", os.Getenv("GT_U_TOKEN"))
	flag.Set("p-token", os.Getenv("GT_P_TOKEN"))
	flag.Set("p-id", os.Getenv("GT_P_ID"))
	flag.Set("url", os.Getenv("GT_URL"))
	flag.Set("v", os.Getenv("GT_V_1"))
	flag.Set("v", os.Getenv("GT_V_2"))
	flag.Set("v", os.Getenv("GT_V_3"))
	main()
}
