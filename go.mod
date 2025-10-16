module github.com/rshade/pulumicost-plugin-vantage

go 1.24.7

replace (
	github.com/rshade/pulumicost-core => ../pulumicost-core
	github.com/rshade/pulumicost-spec => ../pulumicost-spec
)

require github.com/spf13/cobra v1.10.1

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)
