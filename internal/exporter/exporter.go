package exporter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/danopstech/starlink_exporter/pkg/spacex.com/api/device"
)

const (
	// DishAddress to reach Starlink dish ip:port
	DishAddress = "192.168.100.1:9200"
	namespace   = "starlink"
)

var (
	dishUp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "up"),
		"Was the last query of Starlink dish successful.",
		nil, nil,
	)
	dishScrapeDurationSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "scrape_duration_seconds"),
		"Time to scrape metrics from starlink dish",
		nil, nil,
	)

	// collectDishContext
	dishInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "info"),
		"Running software versions and IDs of hardware",
		[]string{"device_id", "hardware_version", "software_version", "country_code", "utc_offset"}, nil,
	)
	dishUptimeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "uptime_seconds"),
		"Dish running time",
		nil, nil,
	)
	dishCellId = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "cell_id"),
		"Cell ID dish is located in",
		nil, nil,
	)
	dishPopRackId = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "pop_rack_id"),
		"pop rack id",
		nil, nil,
	)
	dishInitialSatelliteId = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "initial_satellite_id"),
		"initial satellite id",
		nil, nil,
	)
	dishInitialGatewayId = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "initial_gateway_id"),
		"initial gateway id",
		nil, nil,
	)
	dishOnBackupBeam = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "backup_beam"),
		"connected to backup beam",
		nil, nil,
	)
	dishSecondsToSlotEnd = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "time_to_slot_end_seconds"),
		"Seconds left on current slot",
		nil, nil,
	)

	// collectDishStatus
	dishState = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "state"),
		"The current dishState of the Dish (Unknown, Booting, Searching, Connected).",
		nil, nil,
	)
	dishSecondsToFirstNonemptySlot = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "first_nonempty_slot_seconds"),
		"Seconds to next non empty slot",
		nil, nil,
	)
	dishPopPingDropRatio = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "pop_ping_drop_ratio"),
		"Percent of pings dropped",
		nil, nil,
	)
	dishPopPingLatencySeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "pop_ping_latency_seconds"),
		"Latency of connection in seconds",
		nil, nil,
	)
	dishSnr = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "snr"),
		"Signal strength of the connection",
		nil, nil,
	)
	dishUplinkThroughputBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "uplink_throughput_bytes"),
		"Amount of bandwidth in bytes per second upload",
		nil, nil,
	)
	dishDownlinkThroughputBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "downlink_throughput_bytes"),
		"Amount of bandwidth in bytes per second download",
		nil, nil,
	)

	dishBoreSightAzimuthDeg = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "bore_sight_azimuth_deg"),
		"azimuth in degrees",
		nil, nil,
	)

	dishBoreSightElevationDeg = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "bore_sight_elevation_deg"),
		"elevation in degrees",
		nil, nil,
	)

	// collectDishGPS
	dishGpsValid = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "gps_valid"),
		"GPS signal validity status",
		nil, nil,
	)
	dishGpsSats = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "gps_satellites"),
		"Number of GPS satellites in view",
		nil, nil,
	)
	dishNoSatsAfterTtff = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "no_sats_after_ttff"),
		"No satellites after time to first fix",
		nil, nil,
	)
	dishInhibitGps = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "inhibit_gps"),
		"GPS inhibited status",
		nil, nil,
	)

	// collectDishAlignment
	dishTiltAngleDeg = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "tilt_angle_deg"),
		"Tilt angle in degrees",
		nil, nil,
	)
	dishDesiredBoreSightAzimuthDeg = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "desired_bore_sight_azimuth_deg"),
		"Desired azimuth in degrees",
		nil, nil,
	)
	dishDesiredBoreSightElevationDeg = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "desired_bore_sight_elevation_deg"),
		"Desired elevation in degrees",
		nil, nil,
	)
	dishActuatorState = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "actuator_state"),
		"Actuator state",
		nil, nil,
	)
	dishAttitudeEstimationState = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "attitude_estimation_state"),
		"Attitude estimation state",
		nil, nil,
	)
	dishAttitudeUncertaintyDeg = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "attitude_uncertainty_deg"),
		"Attitude uncertainty in degrees",
		nil, nil,
	)

	// collectDishReadyStates
	dishReadyStateCady = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "ready_state_cady"),
		"Cady ready state",
		nil, nil,
	)
	dishReadyStateScp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "ready_state_scp"),
		"SCP ready state",
		nil, nil,
	)
	dishReadyStateL1L2 = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "ready_state_l1l2"),
		"L1L2 ready state",
		nil, nil,
	)
	dishReadyStateXphy = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "ready_state_xphy"),
		"XPHY ready state",
		nil, nil,
	)
	dishReadyStateAap = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "ready_state_aap"),
		"AAP ready state",
		nil, nil,
	)
	dishReadyStateRf = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "ready_state_rf"),
		"RF ready state",
		nil, nil,
	)

	// collectDishSoftwareUpdate
	dishSoftwareUpdateState = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "software_update_state"),
		"Software update state",
		nil, nil,
	)
	dishSoftwareUpdateProgress = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "software_update_progress"),
		"Software update progress (0-1)",
		nil, nil,
	)
	dishUpdateRequiresReboot = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "update_requires_reboot"),
		"Update requires reboot",
		nil, nil,
	)

	// collectDishInitialization
	dishInitAttitudeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "init_attitude_seconds"),
		"Attitude initialization time in seconds",
		nil, nil,
	)
	dishInitBurstDetectedSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "init_burst_detected_seconds"),
		"Burst detection time in seconds",
		nil, nil,
	)
	dishInitEkfConvergedSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "init_ekf_converged_seconds"),
		"EKF convergence time in seconds",
		nil, nil,
	)
	dishInitFirstCplaneSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "init_first_cplane_seconds"),
		"First CPLANE time in seconds",
		nil, nil,
	)
	dishInitFirstPopPingSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "init_first_pop_ping_seconds"),
		"First POP ping time in seconds",
		nil, nil,
	)
	dishInitGpsValidSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "init_gps_valid_seconds"),
		"GPS valid time in seconds",
		nil, nil,
	)
	dishInitNetworkEntrySeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "init_network_entry_seconds"),
		"Network entry time in seconds",
		nil, nil,
	)
	dishInitNetworkScheduleSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "init_network_schedule_seconds"),
		"Network schedule time in seconds",
		nil, nil,
	)
	dishInitRfReadySeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "init_rf_ready_seconds"),
		"RF ready time in seconds",
		nil, nil,
	)
	dishInitStableConnectionSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "init_stable_connection_seconds"),
		"Stable connection time in seconds",
		nil, nil,
	)

	// collectDishObstructions
	dishCurrentlyObstructed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "currently_obstructed"),
		"Status of view of the sky",
		nil, nil,
	)
	dishFractionObstructionRatio = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "fraction_obstruction_ratio"),
		"Percentage of obstruction",
		nil, nil,
	)
	dishLast24hObstructedSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "last_24h_obstructed_seconds"),
		"Number of seconds view of sky has been obstructed in the last 24hours",
		nil, nil,
	)
	dishValidSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "valid_seconds"),
		"Unknown",
		nil, nil,
	)
	dishProlongedObstructionDurationSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "prolonged_obstruction_duration_seconds"),
		"Average in seconds of prolonged obstructions",
		nil, nil,
	)
	dishProlongedObstructionIntervalSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "prolonged_obstruction_interval_seconds"),
		"Average prolonged obstruction interval in seconds",
		nil, nil,
	)
	dishWedgeFractionObstructionRatio = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "wedge_fraction_obstruction_ratio"),
		"Percentage of obstruction per wedge section",
		[]string{"wedge", "wedge_name"}, nil,
	)
	dishWedgeAbsFractionObstructionRatio = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "wedge_abs_fraction_obstruction_ratio"),
		"Percentage of Absolute fraction per wedge section",
		[]string{"wedge", "wedge_name"}, nil,
	)

	// collectDishAlerts
	dishAlertMotorsStuck = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_motors_stuck"),
		"Status of motor stuck",
		nil, nil,
	)
	dishAlertThermalThrottle = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_thermal_throttle"),
		"Status of thermal throttling",
		nil, nil,
	)
	dishAlertThermalShutdown = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_thermal_shutdown"),
		"Status of thermal shutdown",
		nil, nil,
	)
	dishAlertMastNotNearVertical = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_mast_not_near_vertical"),
		"Status of mast position",
		nil, nil,
	)
	dishUnexpectedLocation = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_unexpected_location"),
		"Status of location",
		nil, nil,
	)
	dishSlowEthernetSpeeds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_slow_eth_speeds"),
		"Status of ethernet",
		nil, nil,
	)
	dishAlertSlowEthernetSpeeds100 = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_slow_eth_speeds_100"),
		"Status of 100Mbps ethernet",
		nil, nil,
	)
	dishAlertRoaming = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_roaming"),
		"Status of roaming",
		nil, nil,
	)
	dishAlertInstallPending = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_install_pending"),
		"Installation pending status",
		nil, nil,
	)
	dishAlertIsHeating = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_is_heating"),
		"Heating status",
		nil, nil,
	)
	dishAlertPowerSupplyThermalThrottle = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_power_supply_thermal_throttle"),
		"Power supply thermal throttle status",
		nil, nil,
	)
	dishAlertIsPowerSaveIdle = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_is_power_save_idle"),
		"Power save idle status",
		nil, nil,
	)
	dishAlertMovingTooFast = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dish", "alert_moving_too_fast"),
		"Moving too fast status",
		nil, nil,
	)
)

// Exporter collects Starlink stats from the Dish and exports them using
// the prometheus metrics package.
type Exporter struct {
	Conn        *grpc.ClientConn
	Client      device.DeviceClient
	DishID      string
	CountryCode string
}

// New returns an initialized Exporter.
func New(address string) (*Exporter, error) {
	ctx, connCancel := context.WithTimeout(context.Background(), time.Second*3)
	defer connCancel()
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("error creating underlying gRPC connection to starlink dish: %s", err.Error())
	}

	ctx, HandleCancel := context.WithTimeout(context.Background(), time.Second*1)
	defer HandleCancel()
	resp, err := device.NewDeviceClient(conn).Handle(ctx, &device.Request{
		Request: &device.Request_GetDeviceInfo{},
	})
	if err != nil {
		return nil, fmt.Errorf("could not collect inital information from dish: %s", err.Error())
	}

	return &Exporter{
		Conn:        conn,
		Client:      device.NewDeviceClient(conn),
		DishID:      resp.GetGetDeviceInfo().GetDeviceInfo().GetId(),
		CountryCode: resp.GetGetDeviceInfo().GetDeviceInfo().GetCountryCode(),
	}, nil
}

// Describe describes all the metrics ever exported by the Starlink exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- dishUp
	ch <- dishScrapeDurationSeconds

	// collectDishContext
	ch <- dishInfo
	ch <- dishUptimeSeconds
	ch <- dishCellId
	ch <- dishPopRackId
	ch <- dishInitialSatelliteId
	ch <- dishInitialGatewayId
	ch <- dishOnBackupBeam
	ch <- dishSecondsToSlotEnd

	// collectDishStatus
	ch <- dishState
	ch <- dishSecondsToFirstNonemptySlot
	ch <- dishPopPingDropRatio
	ch <- dishPopPingLatencySeconds
	ch <- dishSnr
	ch <- dishUplinkThroughputBytes
	ch <- dishDownlinkThroughputBytes
	ch <- dishBoreSightAzimuthDeg
	ch <- dishBoreSightElevationDeg

	// collectDishGPS
	ch <- dishGpsValid
	ch <- dishGpsSats
	ch <- dishNoSatsAfterTtff
	ch <- dishInhibitGps

	// collectDishAlignment
	ch <- dishTiltAngleDeg
	ch <- dishDesiredBoreSightAzimuthDeg
	ch <- dishDesiredBoreSightElevationDeg
	ch <- dishActuatorState
	ch <- dishAttitudeEstimationState
	ch <- dishAttitudeUncertaintyDeg

	// collectDishReadyStates
	ch <- dishReadyStateCady
	ch <- dishReadyStateScp
	ch <- dishReadyStateL1L2
	ch <- dishReadyStateXphy
	ch <- dishReadyStateAap
	ch <- dishReadyStateRf

	// collectDishSoftwareUpdate
	ch <- dishSoftwareUpdateState
	ch <- dishSoftwareUpdateProgress
	ch <- dishUpdateRequiresReboot

	// collectDishInitialization
	ch <- dishInitAttitudeSeconds
	ch <- dishInitBurstDetectedSeconds
	ch <- dishInitEkfConvergedSeconds
	ch <- dishInitFirstCplaneSeconds
	ch <- dishInitFirstPopPingSeconds
	ch <- dishInitGpsValidSeconds
	ch <- dishInitNetworkEntrySeconds
	ch <- dishInitNetworkScheduleSeconds
	ch <- dishInitRfReadySeconds
	ch <- dishInitStableConnectionSeconds

	// collectDishObstructions
	ch <- dishCurrentlyObstructed
	ch <- dishFractionObstructionRatio
	ch <- dishLast24hObstructedSeconds
	ch <- dishValidSeconds
	ch <- dishProlongedObstructionDurationSeconds
	ch <- dishProlongedObstructionIntervalSeconds
	ch <- dishWedgeFractionObstructionRatio
	ch <- dishWedgeAbsFractionObstructionRatio

	// collectDishAlerts
	ch <- dishAlertMotorsStuck
	ch <- dishAlertThermalThrottle
	ch <- dishAlertThermalShutdown
	ch <- dishAlertMastNotNearVertical
	ch <- dishUnexpectedLocation
	ch <- dishSlowEthernetSpeeds
	ch <- dishAlertSlowEthernetSpeeds100
	ch <- dishAlertRoaming
	ch <- dishAlertInstallPending
	ch <- dishAlertIsHeating
	ch <- dishAlertPowerSupplyThermalThrottle
	ch <- dishAlertIsPowerSaveIdle
	ch <- dishAlertMovingTooFast
}

// Collect fetches the stats from Starlink dish and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()

	ok := e.collectDishContext(ch)
	ok = ok && e.collectDishStatus(ch)
	ok = ok && e.collectDishGPS(ch)
	ok = ok && e.collectDishAlignment(ch)
	ok = ok && e.collectDishReadyStates(ch)
	ok = ok && e.collectDishSoftwareUpdate(ch)
	ok = ok && e.collectDishInitialization(ch)
	ok = ok && e.collectDishObstructions(ch)
	ok = ok && e.collectDishAlerts(ch)

	if ok {
		ch <- prometheus.MustNewConstMetric(
			dishUp, prometheus.GaugeValue, 1.0,
		)
		ch <- prometheus.MustNewConstMetric(
			dishScrapeDurationSeconds, prometheus.GaugeValue, time.Since(start).Seconds(),
		)
	} else {
		ch <- prometheus.MustNewConstMetric(
			dishUp, prometheus.GaugeValue, 0.0,
		)
	}
}

func (e *Exporter) collectDishContext(ch chan<- prometheus.Metric) bool {
	req := &device.Request{
		Request: &device.Request_DishGetContext{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	resp, err := e.Client.Handle(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() != 7 {
			log.Errorf("failed to collect dish context: %s", err.Error())
			return false
		}
	}

	dishC := resp.GetDishGetContext()
	dishI := dishC.GetDeviceInfo()
	dishS := dishC.GetDeviceState()

	ch <- prometheus.MustNewConstMetric(
		dishInfo, prometheus.GaugeValue, 1.00,
		dishI.GetId(),
		dishI.GetHardwareVersion(),
		dishI.GetSoftwareVersion(),
		dishI.GetCountryCode(),
		fmt.Sprint(dishI.GetUtcOffsetS()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishUptimeSeconds, prometheus.GaugeValue, float64(dishS.GetUptimeS()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishCellId, prometheus.GaugeValue, float64(dishC.GetCellId()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishPopRackId, prometheus.GaugeValue, float64(dishC.GetPopRackId()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishInitialSatelliteId, prometheus.GaugeValue, float64(dishC.GetInitialSatelliteId()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishInitialGatewayId, prometheus.GaugeValue, float64(dishC.GetInitialGatewayId()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishOnBackupBeam, prometheus.GaugeValue, flool(dishC.GetOnBackupBeam()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishSecondsToSlotEnd, prometheus.GaugeValue, float64(dishC.GetSecondsToSlotEnd()),
	)

	return true
}

func (e *Exporter) collectDishStatus(ch chan<- prometheus.Metric) bool {
	req := &device.Request{
		Request: &device.Request_GetStatus{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	resp, err := e.Client.Handle(ctx, req)
	if err != nil {
		log.Errorf("failed to collect status from dish: %s", err.Error())
		return false
	}

	dishStatus := resp.GetDishGetStatus()

	ch <- prometheus.MustNewConstMetric(
		dishState, prometheus.GaugeValue, float64(dishStatus.GetState().Number()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishSecondsToFirstNonemptySlot, prometheus.GaugeValue, float64(dishStatus.GetSecondsToFirstNonemptySlot()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishPopPingDropRatio, prometheus.GaugeValue, float64(dishStatus.GetPopPingDropRate()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishPopPingLatencySeconds, prometheus.GaugeValue, float64(dishStatus.GetPopPingLatencyMs()/1000),
	)

	ch <- prometheus.MustNewConstMetric(
		dishSnr, prometheus.GaugeValue, float64(dishStatus.GetSnr()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishUplinkThroughputBytes, prometheus.GaugeValue, float64(dishStatus.GetUplinkThroughputBps()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishDownlinkThroughputBytes, prometheus.GaugeValue, float64(dishStatus.GetDownlinkThroughputBps()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishBoreSightAzimuthDeg, prometheus.GaugeValue, float64(dishStatus.GetBoresightAzimuthDeg()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishBoreSightElevationDeg, prometheus.GaugeValue, float64(dishStatus.GetBoresightElevationDeg()),
	)

	return true
}

func (e *Exporter) collectDishObstructions(ch chan<- prometheus.Metric) bool {
	req := &device.Request{
		Request: &device.Request_GetStatus{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	resp, err := e.Client.Handle(ctx, req)
	if err != nil {
		log.Errorf("failed to collect obstructions from dish: %s", err.Error())
		return false
	}

	obstructions := resp.GetDishGetStatus().GetObstructionStats()

	ch <- prometheus.MustNewConstMetric(
		dishCurrentlyObstructed, prometheus.GaugeValue, flool(obstructions.GetCurrentlyObstructed()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishFractionObstructionRatio, prometheus.GaugeValue, float64(obstructions.GetFractionObstructed()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishLast24hObstructedSeconds, prometheus.GaugeValue, float64(obstructions.GetLast_24HObstructedS()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishValidSeconds, prometheus.GaugeValue, float64(obstructions.GetValidS()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishProlongedObstructionDurationSeconds, prometheus.GaugeValue, float64(obstructions.GetAvgProlongedObstructionDurationS()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishProlongedObstructionIntervalSeconds, prometheus.GaugeValue, float64(obstructions.GetAvgProlongedObstructionIntervalS()),
	)

	for i, v := range obstructions.GetWedgeFractionObstructed() {
		ch <- prometheus.MustNewConstMetric(
			dishWedgeFractionObstructionRatio, prometheus.GaugeValue, float64(v),
			strconv.Itoa(i),
			fmt.Sprintf("%d_to_%d", i*30, (i+1)*30),
		)
	}

	for i, v := range obstructions.GetWedgeAbsFractionObstructed() {
		ch <- prometheus.MustNewConstMetric(
			dishWedgeAbsFractionObstructionRatio, prometheus.GaugeValue, float64(v),
			strconv.Itoa(i),
			fmt.Sprintf("%d_to_%d", i*30, (i+1)*30),
		)
	}

	return true
}

func (e *Exporter) collectDishGPS(ch chan<- prometheus.Metric) bool {
	req := &device.Request{
		Request: &device.Request_GetStatus{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	resp, err := e.Client.Handle(ctx, req)
	if err != nil {
		log.Errorf("failed to collect GPS stats from dish: %s", err.Error())
		return false
	}

	gpsStats := resp.GetDishGetStatus().GetGpsStats()

	ch <- prometheus.MustNewConstMetric(
		dishGpsValid, prometheus.GaugeValue, flool(gpsStats.GetGpsValid()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishGpsSats, prometheus.GaugeValue, float64(gpsStats.GetGpsSats()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishNoSatsAfterTtff, prometheus.GaugeValue, flool(gpsStats.GetNoSatsAfterTtff()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishInhibitGps, prometheus.GaugeValue, flool(gpsStats.GetInhibitGps()),
	)

	return true
}

func (e *Exporter) collectDishAlignment(ch chan<- prometheus.Metric) bool {
	req := &device.Request{
		Request: &device.Request_GetStatus{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	resp, err := e.Client.Handle(ctx, req)
	if err != nil {
		log.Errorf("failed to collect alignment stats from dish: %s", err.Error())
		return false
	}

	alignmentStats := resp.GetDishGetStatus().GetAlignmentStats()

	ch <- prometheus.MustNewConstMetric(
		dishTiltAngleDeg, prometheus.GaugeValue, float64(alignmentStats.GetTiltAngleDeg()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishDesiredBoreSightAzimuthDeg, prometheus.GaugeValue, float64(alignmentStats.GetDesiredBoresightAzimuthDeg()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishDesiredBoreSightElevationDeg, prometheus.GaugeValue, float64(alignmentStats.GetDesiredBoresightElevationDeg()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishActuatorState, prometheus.GaugeValue, float64(alignmentStats.GetActuatorState()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishAttitudeEstimationState, prometheus.GaugeValue, float64(alignmentStats.GetAttitudeEstimationState()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishAttitudeUncertaintyDeg, prometheus.GaugeValue, float64(alignmentStats.GetAttitudeUncertaintyDeg()),
	)

	return true
}

func (e *Exporter) collectDishReadyStates(ch chan<- prometheus.Metric) bool {
	req := &device.Request{
		Request: &device.Request_GetStatus{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	resp, err := e.Client.Handle(ctx, req)
	if err != nil {
		log.Errorf("failed to collect ready states from dish: %s", err.Error())
		return false
	}

	readyStates := resp.GetDishGetStatus().GetReadyStates()

	ch <- prometheus.MustNewConstMetric(
		dishReadyStateCady, prometheus.GaugeValue, flool(readyStates.GetCady()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishReadyStateScp, prometheus.GaugeValue, flool(readyStates.GetScp()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishReadyStateL1L2, prometheus.GaugeValue, flool(readyStates.GetL1L2()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishReadyStateXphy, prometheus.GaugeValue, flool(readyStates.GetXphy()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishReadyStateAap, prometheus.GaugeValue, flool(readyStates.GetAap()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishReadyStateRf, prometheus.GaugeValue, flool(readyStates.GetRf()),
	)

	return true
}

func (e *Exporter) collectDishSoftwareUpdate(ch chan<- prometheus.Metric) bool {
	req := &device.Request{
		Request: &device.Request_GetStatus{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	resp, err := e.Client.Handle(ctx, req)
	if err != nil {
		log.Errorf("failed to collect software update stats from dish: %s", err.Error())
		return false
	}

	swStats := resp.GetDishGetStatus().GetSoftwareUpdateStats()

	ch <- prometheus.MustNewConstMetric(
		dishSoftwareUpdateState, prometheus.GaugeValue, float64(swStats.GetSoftwareUpdateState()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishSoftwareUpdateProgress, prometheus.GaugeValue, float64(swStats.GetSoftwareUpdateProgress()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishUpdateRequiresReboot, prometheus.GaugeValue, flool(swStats.GetUpdateRequiresReboot()),
	)

	return true
}

func (e *Exporter) collectDishInitialization(ch chan<- prometheus.Metric) bool {
	req := &device.Request{
		Request: &device.Request_GetStatus{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	resp, err := e.Client.Handle(ctx, req)
	if err != nil {
		log.Errorf("failed to collect initialization stats from dish: %s", err.Error())
		return false
	}

	initStats := resp.GetDishGetStatus().GetInitializationDurationSeconds()

	ch <- prometheus.MustNewConstMetric(
		dishInitAttitudeSeconds, prometheus.GaugeValue, float64(initStats.GetAttitudeInitialization()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishInitBurstDetectedSeconds, prometheus.GaugeValue, float64(initStats.GetBurstDetected()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishInitEkfConvergedSeconds, prometheus.GaugeValue, float64(initStats.GetEkfConverged()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishInitFirstCplaneSeconds, prometheus.GaugeValue, float64(initStats.GetFirstCplane()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishInitFirstPopPingSeconds, prometheus.GaugeValue, float64(initStats.GetFirstPopPing()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishInitGpsValidSeconds, prometheus.GaugeValue, float64(initStats.GetGpsValid()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishInitNetworkEntrySeconds, prometheus.GaugeValue, float64(initStats.GetInitialNetworkEntry()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishInitNetworkScheduleSeconds, prometheus.GaugeValue, float64(initStats.GetNetworkSchedule()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishInitRfReadySeconds, prometheus.GaugeValue, float64(initStats.GetRfReady()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishInitStableConnectionSeconds, prometheus.GaugeValue, float64(initStats.GetStableConnection()),
	)

	return true
}

func (e *Exporter) collectDishAlerts(ch chan<- prometheus.Metric) bool {
	req := &device.Request{
		Request: &device.Request_GetStatus{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	resp, err := e.Client.Handle(ctx, req)
	if err != nil {
		log.Errorf("failed to collect alerts from dish: %s", err.Error())
		return false
	}

	alerts := resp.GetDishGetStatus().GetAlerts()

	ch <- prometheus.MustNewConstMetric(
		dishAlertMotorsStuck, prometheus.GaugeValue, flool(alerts.GetMotorsStuck()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishAlertThermalThrottle, prometheus.GaugeValue, flool(alerts.GetThermalThrottle()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishAlertThermalShutdown, prometheus.GaugeValue, flool(alerts.GetThermalShutdown()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishAlertMastNotNearVertical, prometheus.GaugeValue, flool(alerts.GetMastNotNearVertical()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishUnexpectedLocation, prometheus.GaugeValue, flool(alerts.GetUnexpectedLocation()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishSlowEthernetSpeeds, prometheus.GaugeValue, flool(alerts.GetSlowEthernetSpeeds()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishAlertSlowEthernetSpeeds100, prometheus.GaugeValue, flool(alerts.GetSlowEthernetSpeeds100()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishAlertRoaming, prometheus.GaugeValue, flool(alerts.GetRoaming()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishAlertInstallPending, prometheus.GaugeValue, flool(alerts.GetInstallPending()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishAlertIsHeating, prometheus.GaugeValue, flool(alerts.GetIsHeating()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishAlertPowerSupplyThermalThrottle, prometheus.GaugeValue, flool(alerts.GetPowerSupplyThermalThrottle()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishAlertIsPowerSaveIdle, prometheus.GaugeValue, flool(alerts.GetIsPowerSaveIdle()),
	)

	ch <- prometheus.MustNewConstMetric(
		dishAlertMovingTooFast, prometheus.GaugeValue, flool(alerts.GetMovingTooFastForPolicy()),
	)

	return true
}

func flool(b bool) float64 {
	if b {
		return 1.00
	}
	return 0.00
}
