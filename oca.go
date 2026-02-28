package oca

type OcaONo uint32

type OcaNotificationDeliveryMode uint8

func (m OcaNotificationDeliveryMode) MarshalBinary() ([]byte, error) {
	return []byte{byte(m)}, nil
}

const (
	OcaNotificationDeliveryModeNormal      OcaNotificationDeliveryMode = 1
	OcaNotificationDeliveryModeLightweight OcaNotificationDeliveryMode = 2
)

const (
	OcaDeviceManager          OcaONo = 1
	OcaSecurityManager        OcaONo = 2
	OcaFirmwareManager        OcaONo = 3
	OcaSubscriptionManager    OcaONo = 4
	OcaPowerManager           OcaONo = 5
	OcaNetworkManager         OcaONo = 6
	OcaMediaClockManager      OcaONo = 7
	OcaAudioProcessingManager OcaONo = 9
	OcaDeviceTimeManager      OcaONo = 10
	OcaDiagnosticManager      OcaONo = 13
	OcaLockManager            OcaONo = 14
	OcaBlockManager           OcaONo = 100
)
