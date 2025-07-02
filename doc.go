/*
Package sped-nfe-go provides functionality for Brazilian Electronic Invoice (NFe)
generation, digital signature, and transmission to SEFAZ webservices.

This package is a Go implementation inspired by the PHP nfephp-org/sped-nfe project,
offering improved performance, type safety, and native concurrency support.

Basic usage:

	config := nfe.Config{
		Environment: nfe.Homologation,
		UF:          nfe.SP,
		Timeout:     30,
	}

	client, err := nfe.New(config)
	if err != nil {
		log.Fatal(err)
	}

	// Generate access key
	accessKey := client.GenerateAccessKey("12345678000190", 55, 1, 123456, 1)

For more examples, see the examples/ directory in the repository.
*/
package spednfego
