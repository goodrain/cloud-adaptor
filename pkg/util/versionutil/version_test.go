package versionutil

import "testing"

func TestCheckVersion(t *testing.T) {
	tests := []struct {
		version string
		isPass  bool
	}{
		{
			version: "1.15.9",
			isPass:  false,
		},
		{
			version: "1.16.0",
			isPass:  true,
		},
		{
			version: "1.19.0",
			isPass:  true,
		},
		{
			version: "1.20.0",
			isPass:  false,
		},
		{
			version: "v1.15.9-rke",
			isPass:  false,
		},
		{
			version: "v1.16.0-rke",
			isPass:  true,
		},
		{
			version: "v1.19.0-rke",
			isPass:  true,
		},
		{
			version: "v1.20.0-rke",
			isPass:  false,
		},
		{
			version: "1.15.9-aliyun.1",
			isPass:  false,
		},
		{
			version: "1.16.0-aliyun.1",
			isPass:  true,
		},
		{
			version: "1.19.0-aliyun.1",
			isPass:  true,
		},
		{
			version: "1.20.0-aliyun.1",
			isPass:  false,
		},
	}
	for _,tc:= range tests{
		if CheckVersion(tc.version) != tc.isPass{
			t.Fatalf("version %v expect %v, actual %v", tc.version, tc.isPass, CheckVersion(tc.version))
		}
	}
}
