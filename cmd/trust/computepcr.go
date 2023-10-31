package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/project-machine/mos/pkg/trust"
	"github.com/project-machine/mos/pkg/utils"
	"github.com/urfave/cli"
)

var computePCR7Cmd = cli.Command{
	Name:      "computePCR7",
	Usage:     "Compute PCR7 value for a given keyset",
	ArgsUsage: "<keyset-name>",
	Action:    doComputePCR7,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "output",
			Usage: "save pcr7 hash values to a file",
		},
		cli.BoolFlag{
			Name:  "events",
			Usage: "Show the list of pcr7 events and their calculated hashes",
		},
	},
}

func doComputePCR7(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) != 1 {
		return errors.New("Required argument: keysetName")
	}
	keysetName := args[0]
	if keysetName == "" {
		return errors.New("Please specify a keyset name")
	}

	trustDir, err := utils.GetMosKeyPath()
	if err != nil {
		return err
	}
	keysetPath := filepath.Join(trustDir, keysetName)
	if !utils.PathExists(keysetPath) {
		return fmt.Errorf("Keyset not found: %s", keysetName)
	}

	pData, err := trust.ComputePCR7(keysetName)
	if err != nil {
		return fmt.Errorf("Failed to generate pcr7 values for %s keyset: (%w)\n", keysetName, err)
	}

	outFile := ctx.String("output")
	doEvents := ctx.Bool("events")
	var p bytes.Buffer

	if doEvents {
		_, err = p.Write(pData.Pcr7Events)
		if err != nil {
			return fmt.Errorf("Failed writing to buffer: (%w)\n", err)
		}
	}
	fmt.Fprintf(&p, "uki-production: %x\n", pData.Pcr7Production)
	fmt.Fprintf(&p, "uki-limited: %x\n", pData.Pcr7Limited)
	fmt.Fprintf(&p, "uki-tpm: %x\n", pData.Pcr7Tpm)

	if outFile != "" {
		err = os.WriteFile(outFile, p.Bytes(), 0640)
		if err != nil {
			return fmt.Errorf("Failed writing to file %s: (%w)\n", outFile, err)
		}
	} else {
		fmt.Printf("%s", p.Bytes())
	}

	return nil
}
