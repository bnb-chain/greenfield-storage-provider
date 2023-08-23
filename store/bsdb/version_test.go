package bsdb

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSpVersion(t *testing.T) {
	// Set the global variables to known values for the test
	AppVersion = "1.0.0"
	GitCommit = "abcd1234"
	GitCommitDate = "2023-08-17"

	// Create a BsDBImpl instance
	db := &BsDBImpl{}

	// Call GetSpVersion
	version := db.GetSpVersion()

	// Check the returned values
	assert.Equal(t, AppVersion, version.SpCodeVersion, "unexpected SpCodeVersion")
	assert.Equal(t, GitCommit, version.SpCodeCommit, "unexpected SpCodeCommit")
	assert.Equal(t, runtime.Version(), version.SpGoVersion, "unexpected SpGoVersion")
	assert.Equal(t, runtime.GOARCH, version.SpArchitecture, "unexpected SpArchitecture")
	assert.Equal(t, runtime.GOOS, version.SpOperatingSystem, "unexpected SpOperatingSystem")
}
