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
