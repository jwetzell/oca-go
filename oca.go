package oca

type OcaONo uint32

type OcaNotificationDeliveryMode uint8

const (
	OcaNotificationDeliveryModeNormal      OcaNotificationDeliveryMode = 1
	OcaNotificationDeliveryModeLightweight OcaNotificationDeliveryMode = 2
)
