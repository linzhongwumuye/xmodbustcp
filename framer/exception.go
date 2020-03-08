package framer

import "fmt"

type Exception uint8

var (
	Success Exception
	IllegalFunction Exception = 1
	IllegalDataAddress Exception = 2
	IllegalDataValue Exception = 3
	SlaveDeviceFailure Exception = 4
	AcknowledgeSlave Exception = 5
	SlaveDeviceBusy Exception = 6
	NegativeAcknowledge Exception = 7
	MemoryParityError Exception = 8
	GatewayPathUnavailable Exception = 10
	GatewayTargetDeviceFailedtoRespond Exception = 11
)

func (e Exception) Error() string {
	return fmt.Sprintf("%d", e)
}

func (e Exception) String() string {
	var str string
	switch e {
	case Success:
		str = fmt.Sprintf("Success")
	case IllegalFunction:
		str = fmt.Sprintf("IllegalFunction")
	case IllegalDataAddress:
		str = fmt.Sprintf("IllegalDataAddress")
	case IllegalDataValue:
		str = fmt.Sprintf("IllegalDataValue")
	case SlaveDeviceFailure:
		str = fmt.Sprintf("SlaveDeviceFailure")
	case AcknowledgeSlave:
		str = fmt.Sprintf("AcknowledgeSlave")
	case SlaveDeviceBusy:
		str = fmt.Sprintf("SlaveDeviceBusy")
	case NegativeAcknowledge:
		str = fmt.Sprintf("NegativeAcknowledge")
	case MemoryParityError:
		str = fmt.Sprintf("MemoryParityError")
	case GatewayPathUnavailable:
		str = fmt.Sprintf("GatewayPathUnavailable")
	case GatewayTargetDeviceFailedtoRespond:
		str = fmt.Sprintf("GatewayTargetDeviceFailedtoRespond")
	default:
		str = fmt.Sprintf("unknown")
	}
	return str
}
