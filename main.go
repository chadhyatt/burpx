package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

var (
	inFile = flag.String("i", "sitemap.xml", "Path to the Burp Suite XML sitemap export to extract")
	outDir = flag.String("o", ".", "Base directory for extracted sitemap")

	shouldWriteDups = flag.Bool("dup", false, "Writes file paths with duplicate response entries with an _n suffix")
)

func usage(exitCode int) {
	flag.Usage()
	os.Exit(exitCode)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "USAGE: %s [OPTION]...\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *inFile == "sitemap.xml" && !PathExists(*inFile) {
		fmt.Fprintf(os.Stderr, "Must provide -i or save the export from Burp as \"sitemap.xml\". See usage:\n\n")
		usage(1)
	}

	rt := new(Root)

	f, err := os.Open(*inFile)
	if err != nil {
		slog.Error("Failed to read export file", "err", err)
		os.Exit(1)
	}

	slog.Info("Deserializing export file")
	if err := rt.Deserialize(f); err != nil {
		_ = f.Close()
		slog.Error("Failed to deserialize export file", "err", err)
		os.Exit(1)
	}

	_ = f.Close()

	slog.Info("Extracting export to directory", "burpVersion", rt.BurpVersion, "exportTime", rt.ExportTime, "out", *outDir)
	if err := ExtractItems(rt, *outDir); err != nil {
		slog.Error("Error extracting export file", "err", err)
		os.Exit(1)
	}
}
