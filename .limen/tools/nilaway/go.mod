// nilaway's own module: its dependency graph stays exactly its upstream's,
// which a shared tools module could not promise (book/tooling.md).
module github.com/forkcloser/godolint/tools/nilaway

go 1.26.3

tool go.uber.org/nilaway/cmd/nilaway

require (
	github.com/klauspost/compress v1.20.0 // indirect
	go.uber.org/nilaway v0.0.0-20260918162853-acb8859b9031 // indirect
	golang.org/x/exp/typeparams v0.0.0-20260908205506-85c1c2202aba // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/tools v0.51.0 // indirect
)
