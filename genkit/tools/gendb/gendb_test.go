package main

import "testing"

func Test_processPackage(t *testing.T) {
	type args struct {
		pkgPath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test_processPackage",
			args: args{
				pkgPath: "/Users/xq.yan/goprojects/agent-demo/tools/sqlconstgen/test_model.go",
			},
		},
		{
			name: "Test directory",
			args: args{
				pkgPath: "/Users/xq.yan/goprojects/agent-demo/tools/sqlconstgen/test_model",
			},
		},
		{
			name: "Test_processPackage",
			args: args{
				pkgPath: "/Users/xq.yan/goprojects/agent-demo/apps/apiserver/internal/domain",
			},
		},
		{
			name: "Test directory",
			args: args{
				pkgPath: "/Users/xq.yan/goprojects/agent-demo/tools/sqlconstgen/test_model",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processPackage(tt.args.pkgPath)
		})
	}
}
