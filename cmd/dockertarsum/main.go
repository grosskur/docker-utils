package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vbatts/docker-utils/opts"
	"github.com/vbatts/docker-utils/sum"
	"github.com/vbatts/docker-utils/version"
)

func main() {
	var (
		checks       = sum.Checks{}
		failedChecks = []bool{}
	)
	flag.Parse()

	if *flVersion {
		fmt.Printf("%s - %s\n", os.Args[0], version.VERSION)
		os.Exit(0)
	}

	if len(flChecks.Args) > 0 {
		for _, c := range flChecks.Args {
			fh, err := os.Open(c)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				os.Exit(1)
			}
			newChecks, err := sum.ReadChecks(fh)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				os.Exit(1)
			}
			checks = append(checks, newChecks...)
		}
	}

	if flag.NArg() == 0 {
		if *flStream {
			var hashes map[string]string
			var err error
			if !*flRootTar {
				// assumption is this is stdin from `docker save`
				if hashes, err = sum.SumAllDockerSave(os.Stdin); err != nil {
					fmt.Printf("ERROR: %s\n", err)
					os.Exit(1)
				}
			} else {
				hash, err := sum.SumTarLayer(os.Stdin, nil, nil)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err)
					os.Exit(1)
				}
				hashes = map[string]string{"-": hash}
			}
			for id, hash := range hashes {
				if len(checks) == 0 {
					// nothing to check against, just print the hash
					fmt.Printf("%s%s-:%s\n", hash, sum.DefaultSpacer, id)
				} else {
					// check the sum against the checks available
					check := checks.Get(id)
					if check == nil {
						fmt.Fprintf(os.Stderr, "WARNING: no check found for ID [%s]\n", id)
						continue
					}
					check.Seen = true // so can print NOT FOUND IDs
					var result string
					if check.Hash != hash {
						result = "FAILED"
						failedChecks = append(failedChecks, false)
					} else {
						result = "OK"
					}
					fmt.Printf("%s:%s%s\n", id, sum.DefaultSpacer, result)
				}
			}
		} else {
			// maybe the actual layer.tar ? and json? or image name and we'll call a docker daemon?
			fmt.Println("ERROR: not implemented yet")
			os.Exit(2)
		}
	}

	for _, arg := range flag.Args() {
		fh, err := os.Open(arg)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
		if *flStream {
			var hashes map[string]string
			if !*flRootTar {
				// assumption is this is a tar from `docker save`
				if hashes, err = sum.SumAllDockerSave(fh); err != nil {
					fmt.Printf("ERROR: %s\n", err)
					os.Exit(1)
				}
			} else {
				hash, err := sum.SumTarLayer(fh, nil, nil)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err)
					os.Exit(1)
				}
				hashes = map[string]string{arg: hash}
			}
			for id, hash := range hashes {
				if len(checks) == 0 {
					fmt.Printf("%s%s%s:%s\n", hash, sum.DefaultSpacer, arg, id)
				} else {
					// check the sum against the checks available
					check := checks.Get(id)
					if check == nil {
						fmt.Fprintf(os.Stderr, "WARNING: no check found for ID [%s]\n", id)
						continue
					}
					check.Seen = true // so can print NOT FOUND IDs
					var result string
					if check.Hash != hash {
						result = "FAILED"
						failedChecks = append(failedChecks, false)
					} else {
						result = "OK"
					}
					fmt.Printf("%s:%s%s\n", id, sum.DefaultSpacer, result)
				}
			}
		} else {
			// maybe the actual layer.tar ? and json? or image name and we'll call a docker daemon?
			fmt.Println("ERROR: not implemented yet")
			os.Exit(2)
		}
	}

	// print out the rest of the checks info
	if len(checks) > 0 {
		for _, c := range checks {
			if !c.Seen {
				fmt.Printf("%s:%sNOT FOUND\n", c.Id, sum.DefaultSpacer)
			}
		}
		if len(failedChecks) > 0 {
			fmt.Printf("%s: WARNING: %d computed checksums did NOT match\n", os.Args[0], len(failedChecks))
			os.Exit(1)
		}
	}
}

var (
	flChecks  = opts.List{}
	flStream  = flag.Bool("s", true, "read FILEs (or stdin) as the output of `docker save` (this is default)")
	flVersion = flag.Bool("v", false, "show version")
	flRootTar = flag.Bool("r", false, "treat the tar(s) root filesystem archives (not a tar of layers)")
)

func init() {
	flag.Var(&flChecks, "c", "read TarSums from the FILEs (or stdin) and check them")
}
