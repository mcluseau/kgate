package main

import (
	"archive/zip"
	"log"
	"os"

	k "github.com/mcluseau/kubeclient"
)

func genKeyCommand() *Command {
	cmd := &Command{
		Use: "gen-key",
		Run: genKeyRun,
	}

	return cmd
}

func genKeyRun(cmd *Command, args []string) {
	secCA, err := k.Client().CoreV1().Secrets(namespace).Get(secretCA, getOpts)
	if err != nil {
		log.Fatal(err)
	}

	sec := getOrCreateTLS("client", func() ([]byte, []byte) {
		key, keyPEM := PrivateKeyPEM()
		crtPEM := HostCertificatePEM(secCA.Data, 1, key, "client")
		return keyPEM, crtPEM
	})

	zipFile := "kgate-client-config.zip"

	out, err := os.Create(zipFile)
	if err != nil {
		log.Fatal(err)
	}

	defer out.Close()

	log.Print("Writing configuration to ", zipFile)

	zw := zip.NewWriter(out)

	for name, data := range map[string][]byte{
		"url":         []byte("ws://kgate." + namespace + ".dev.isi.nc:80"),
		"server-name": []byte("kgate"),
		"ca.crt":      secCA.Data["tls.crt"],
		"client.crt":  sec.Data["tls.crt"],
		"client.key":  sec.Data["tls.key"],
	} {
		f, err := zw.Create(name)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := f.Write(data); err != nil {
			log.Fatal(err)
		}
	}

	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}
}