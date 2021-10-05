package gasschedules

// TODO: go:embed can be used after we upgrade to go 1.16
// import _ "embed"

// //go:embed gasScheduleV1.toml
// var gasScheduleV1 string

// //go:embed gasScheduleV2.toml
// var gasScheduleV2 string

// //go:embed gasScheduleV3.toml
// var gasScheduleV3 string

//go:generate go run scripts/includetoml.go

// Deprecated: GetV1 yields an old version of the gas schedule.
func GetV1() string {
	return gasScheduleV1
}

// Deprecated: GetV2 yields an old version of the gas schedule.
func GetV2() string {
	return gasScheduleV2
}
func GetV3() string {
	return gasScheduleV3
}
