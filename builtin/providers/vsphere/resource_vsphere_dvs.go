package vsphere

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	//"github.com/vmware/govmomi/vim25/soap"
	"golang.org/x/net/context"
)

type dvs struct {
	datacenter      string
	folder		string
	configSpec	dvsConfigSpec
	ProductInfo 	*distributedVirtualSwitchProductSpec `xml:"productInfo,omitempty"`
	Capability  	*dVSCapability

}

/**
Base three structs for data for dvs
 */
type dvsConfigSpec struct {
	configVersion                       string                                         `xml:"configVersion,omitempty"`
	name                                string                                         `xml:"name,omitempty"`
	numStandalonePorts                  int32                                          `xml:"numStandalonePorts,omitempty"`
	maxPorts                            int32                                          `xml:"maxPorts,omitempty"`
	// need to check wth this is
	//uplinkPortPolicy                    baseDVSUplinkPortPolicy                        `xml:"uplinkPortPolicy,omitempty,typeattr"`
	// Think this is a String array ... but huh??
	uplinkPortgroup                     []managedObjectReference                       `xml:"uplinkPortgroup,omitempty"`

	defaultPortConfig                   baseDVPortSetting                              `xml:"defaultPortConfig,omitempty,typeattr"`
	host                                []distributedVirtualSwitchHostMemberConfigSpec `xml:"host,omitempty"`
	extensionKey                        string                                         `xml:"extensionKey,omitempty"`
	description                         string                                         `xml:"description,omitempty"`
	policy                              *dVSPolicy                                     `xml:"policy,omitempty"`
	vendorSpecificConfig                []distributedVirtualSwitchKeyedOpaqueBlob      `xml:"vendorSpecificConfig,omitempty"`
	contact                             *dVSContactInfo                                `xml:"contact,omitempty"`
	switchIpAddress                     string                                         `xml:"switchIpAddress,omitempty"`
	defaultProxySwitchMaxNumPorts       int32                                          `xml:"defaultProxySwitchMaxNumPorts,omitempty"`
	infrastructureTrafficResourceConfig []dvsHostInfrastructureTrafficResource         `xml:"infrastructureTrafficResourceConfig,omitempty"`
	networkResourceControlVersion       string                                         `xml:"networkResourceControlVersion,omitempty"`
}

type distributedVirtualSwitchProductSpec struct {

	Name            string `xml:"name,omitempty"`
	Vendor          string `xml:"vendor,omitempty"`
	Version         string `xml:"version,omitempty"`
	Build           string `xml:"build,omitempty"`
	ForwardingClass string `xml:"forwardingClass,omitempty"`
	BundleId        string `xml:"bundleId,omitempty"`
	BundleUrl       string `xml:"bundleUrl,omitempty"`
}

type dVSCapability struct {

	DvsOperationSupported              *bool                                     `xml:"dvsOperationSupported"`
	DvPortGroupOperationSupported      *bool                                     `xml:"dvPortGroupOperationSupported"`
	DvPortOperationSupported           *bool                                     `xml:"dvPortOperationSupported"`
	CompatibleHostComponentProductInfo []DistributedVirtualSwitchHostProductSpec `xml:"compatibleHostComponentProductInfo,omitempty"`
	FeaturesSupported                  BaseDVSFeatureCapability                  `xml:"featuresSupported,omitempty,typeattr"`
}

type BaseDVSFeatureCapability interface {
	GetDVSFeatureCapability() *DVSFeatureCapability
}

type baseDVPortSetting struct {
	blocked                 bool
	vmDirectPathGen2Allowed bool
	inShapingPolicy         *dVSTrafficShapingPolicy `xml:"inShapingPolicy,omitempty"`
	outShapingPolicy        *dVSTrafficShapingPolicy `xml:"outShapingPolicy,omitempty"`
	vendorSpecificConfig    *dVSVendorSpecificConfig `xml:"vendorSpecificConfig,omitempty"`
	networkResourcePoolKey  *string            `xml:"networkResourcePoolKey,omitempty"`
	filterPolicy            *dvsFilterPolicy         `xml:"filterPolicy,omitempty"`

}

type dVSTrafficShapingPolicy struct {
	enabled          bool
	averageBandwidth int32 // FIXME this is a Long ... wth should this be??
	peakBandwidth    int32 // See above
	burstSize        int32 // See above
}

type dVSVendorSpecificConfig struct {
	// KeyValue []DistributedVirtualSwitchKeyedOpaqueBlob
	keyvalue []distributedVirtualSwitchKeyedOpaqueBlob
}

// FIXME shoulld be able to reuse the api values
type distributedVirtualSwitchKeyedOpaqueBlob struct {
	key	   string
	opaqueData string `xml:"opaqueData"`
}

type dvsFilterPolicy struct {
	filterConfig []baseDvsFilterConfig
}

type baseDvsFilterConfig struct {
	key        string              `xml:"key,omitempty"`
	agentName  string              `xml:"agentName,omitempty"`
	slotNumber string              `xml:"slotNumber,omitempty"`
	parameters *dvsFilterParameter `xml:"parameters,omitempty"`
	onFailure  string              `xml:"onFailure,omitempty"`
}

type dvsFilterParameter struct {
	parameters	[]string
}


type managedObjectReference struct {
	Type string
	value string

}

type distributedVirtualSwitchHostMemberConfigSpec struct {
	operation            string                                        `xml:"operation"`
	host                 managedObjectReference                        `xml:"host"`
	// FIXME Dynamic Data again
	// backing              baseDistributedVirtualSwitchHostMemberBacking `xml:"backing,omitempty,typeattr"`
	maxProxySwitchPorts  int32                                         `xml:"maxProxySwitchPorts,omitempty"`
	vendorSpecificConfig []distributedVirtualSwitchKeyedOpaqueBlob     `xml:"vendorSpecificConfig,omitempty"`
}

type dVSPolicy struct {
	autoPreInstallAllowed *bool `xml:"autoPreInstallAllowed"`
	autoUpgradeAllowed    *bool `xml:"autoUpgradeAllowed"`
	partialUpgradeAllowed *bool `xml:"partialUpgradeAllowed"`
}

type dVSContactInfo struct {
	name    string `xml:"name,omitempty"`
	contact string `xml:"contact,omitempty"`
}

type dvsHostInfrastructureTrafficResource struct {
	key            string                                         `xml:"key"`
	description    string                                         `xml:"description,omitempty"`
	allocationInfo dvsHostInfrastructureTrafficResourceAllocation `xml:"allocationInfo"`
}

type dvsHostInfrastructureTrafficResourceAllocation struct {
	limit       int64       `xml:"limit,omitempty"`
	shares      *sharesInfo `xml:"shares,omitempty"`
	reservation int64       `xml:"reservation,omitempty"`
}

type sharesInfo struct {

	shares int32    `xml:"shares"`
	level  string 	`xml:"level"`
}

type VMwareDVSConfigSpec struct {

	PvlanConfigSpec             []VMwareDVSPvlanConfigSpec   `xml:"pvlanConfigSpec,omitempty"`
	VspanConfigSpec             []VMwareDVSVspanConfigSpec   `xml:"vspanConfigSpec,omitempty"`
	MaxMtu                      int32                        `xml:"maxMtu,omitempty"`
	LinkDiscoveryProtocolConfig *LinkDiscoveryProtocolConfig `xml:"linkDiscoveryProtocolConfig,omitempty"`
	IpfixConfig                 *VMwareIpfixConfig           `xml:"ipfixConfig,omitempty"`
	LacpApiVersion              string                       `xml:"lacpApiVersion,omitempty"`
	MulticastFilteringMode      string                       `xml:"multicastFilteringMode,omitempty"`
}

type VMwareDVSPvlanConfigSpec struct {

	PvlanEntry VMwareDVSPvlanMapEntry `xml:"pvlanEntry"`
	Operation  string                 `xml:"operation"`
}

type VMwareDVSPvlanMapEntry struct {

	PrimaryVlanId   int32  `xml:"primaryVlanId"`
	SecondaryVlanId int32  `xml:"secondaryVlanId"`
	PvlanType       string `xml:"pvlanType"`
}
type VMwareDVSVspanConfigSpec struct {

	VspanSession VMwareVspanSession `xml:"vspanSession"`
	Operation    string             `xml:"operation"`
}

type VMwareVspanSession struct {

	Key                   string           `xml:"key,omitempty"`
	Name                  string           `xml:"name,omitempty"`
	Description           string           `xml:"description,omitempty"`
	Enabled               bool             `xml:"enabled"`
	SourcePortTransmitted *VMwareVspanPort `xml:"sourcePortTransmitted,omitempty"`
	SourcePortReceived    *VMwareVspanPort `xml:"sourcePortReceived,omitempty"`
	DestinationPort       *VMwareVspanPort `xml:"destinationPort,omitempty"`
	EncapsulationVlanId   int32            `xml:"encapsulationVlanId,omitempty"`
	StripOriginalVlan     bool             `xml:"stripOriginalVlan"`
	MirroredPacketLength  int32            `xml:"mirroredPacketLength,omitempty"`
	NormalTrafficAllowed  bool             `xml:"normalTrafficAllowed"`
	SessionType           string           `xml:"sessionType,omitempty"`
	SamplingRate          int32            `xml:"samplingRate,omitempty"`
}

type VMwareVspanPort struct {

	PortKey                   []string `xml:"portKey,omitempty"`
	UplinkPortName            []string `xml:"uplinkPortName,omitempty"`
	WildcardPortConnecteeType []string `xml:"wildcardPortConnecteeType,omitempty"`
	Vlans                     []int32  `xml:"vlans,omitempty"`
	IpAddress                 []string `xml:"ipAddress,omitempty"`
}

type LinkDiscoveryProtocolConfig struct {

	Protocol  string `xml:"protocol"`
	Operation string `xml:"operation"`
}

type VMwareIpfixConfig struct {

	CollectorIpAddress  string `xml:"collectorIpAddress,omitempty"`
	CollectorPort       int32  `xml:"collectorPort,omitempty"`
	ObservationDomainId int64  `xml:"observationDomainId,omitempty"`
	ActiveFlowTimeout   int32  `xml:"activeFlowTimeout"`
	IdleFlowTimeout     int32  `xml:"idleFlowTimeout"`
	SamplingRate        int32  `xml:"samplingRate"`
	InternalFlowsOnly   bool   `xml:"internalFlowsOnly"`
}

type DistributedVirtualSwitchHostProductSpec struct {

	ProductLineId string `xml:"productLineId,omitempty"`
	Version       string `xml:"version,omitempty"`
}

type DVSFeatureCapability struct {

	NetworkResourceManagementSupported  bool                                    `xml:"networkResourceManagementSupported"`
	VmDirectPathGen2Supported           bool                                    `xml:"vmDirectPathGen2Supported"`
	NicTeamingPolicy                    []string                                `xml:"nicTeamingPolicy,omitempty"`
	NetworkResourcePoolHighShareValue   int32                                   `xml:"networkResourcePoolHighShareValue,omitempty"`
	NetworkResourceManagementCapability *DVSNetworkResourceManagementCapability `xml:"networkResourceManagementCapability,omitempty"`
	// DynamicData again ...
	//HealthCheckCapability               BaseDVSHealthCheckCapability            `xml:"healthCheckCapability,omitempty,typeattr"`
	RollbackCapability                  *DVSRollbackCapability                  `xml:"rollbackCapability,omitempty"`
	BackupRestoreCapability             *DVSBackupRestoreCapability             `xml:"backupRestoreCapability,omitempty"`
	NetworkFilterSupported              *bool                                   `xml:"networkFilterSupported"`
}

type DVSNetworkResourceManagementCapability struct {

	NetworkResourceManagementSupported       bool  `xml:"networkResourceManagementSupported"`
	NetworkResourcePoolHighShareValue        int32 `xml:"networkResourcePoolHighShareValue"`
	QosSupported                             bool  `xml:"qosSupported"`
	UserDefinedNetworkResourcePoolsSupported bool  `xml:"userDefinedNetworkResourcePoolsSupported"`
	NetworkResourceControlVersion3Supported  *bool `xml:"networkResourceControlVersion3Supported"`
}

type DVSRollbackCapability struct {
	RollbackSupported bool `xml:"rollbackSupported"`
}

type DVSBackupRestoreCapability struct {
	BackupRestoreSupported bool `xml:"backupRestoreSupported"`
}

/*
We need this
ConfigSpec  BaseDVSConfigSpec                    `xml:"configSpec,typeattr"`
	ProductInfo *DistributedVirtualSwitchProductSpec `xml:"productInfo,omitempty"`
	Capability  *DVSCapability                       `xml:"capability,omitempty"`
 */


func resourceVSphereDvs() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereDvsCreate,
		Read:   resourceVSphereDvsRead,
		Update: resourceVSphereDvsUpdate,
		Delete: resourceVSphereDvsDelete,

		Schema: map[string]*schema.Schema{
			"datacenter": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"folder": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

		},
	}
}

func resourceVSphereDvsCreate(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] creating file: %#v", d)
	client := meta.(*govmomi.Client)

	f := file{}

	if v, ok := d.GetOk("datacenter"); ok {
		f.datacenter = v.(string)
	}

	if v, ok := d.GetOk("datastore"); ok {
		f.datastore = v.(string)
	} else {
		return fmt.Errorf("datastore argument is required")
	}

	if v, ok := d.GetOk("source_file"); ok {
		f.sourceFile = v.(string)
	} else {
		return fmt.Errorf("source_file argument is required")
	}

	if v, ok := d.GetOk("destination_file"); ok {
		f.destinationFile = v.(string)
	} else {
		return fmt.Errorf("destination_file argument is required")
	}

	err := createFile(client, &f)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("[%v] %v/%v", f.datastore, f.datacenter, f.destinationFile))
	log.Printf("[INFO] Created file: %s", f.destinationFile)

	return resourceVSphereFileRead(d, meta)
}


func resourceVSphereDvsRead(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] reading file: %#v", d)
	f := file{}

	if v, ok := d.GetOk("datacenter"); ok {
		f.datacenter = v.(string)
	}

	if v, ok := d.GetOk("datastore"); ok {
		f.datastore = v.(string)
	} else {
		return fmt.Errorf("datastore argument is required")
	}

	if v, ok := d.GetOk("source_file"); ok {
		f.sourceFile = v.(string)
	} else {
		return fmt.Errorf("source_file argument is required")
	}

	if v, ok := d.GetOk("destination_file"); ok {
		f.destinationFile = v.(string)
	} else {
		return fmt.Errorf("destination_file argument is required")
	}

	client := meta.(*govmomi.Client)
	finder := find.NewFinder(client.Client, true)

	dc, err := finder.Datacenter(context.TODO(), f.datacenter)
	if err != nil {
		return fmt.Errorf("error %s", err)
	}
	finder = finder.SetDatacenter(dc)

	ds, err := getDatastore(finder, f.datastore)
	if err != nil {
		return fmt.Errorf("error %s", err)
	}

	_, err = ds.Stat(context.TODO(), f.destinationFile)
	if err != nil {
		d.SetId("")
		return err
	}

	return nil
}

func resourceVSphereDvsUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] updating file: %#v", d)
	if d.HasChange("destination_file") {
		oldDestinationFile, newDestinationFile := d.GetChange("destination_file")
		f := file{}

		if v, ok := d.GetOk("datacenter"); ok {
			f.datacenter = v.(string)
		}

		if v, ok := d.GetOk("datastore"); ok {
			f.datastore = v.(string)
		} else {
			return fmt.Errorf("datastore argument is required")
		}

		if v, ok := d.GetOk("source_file"); ok {
			f.sourceFile = v.(string)
		} else {
			return fmt.Errorf("source_file argument is required")
		}

		if v, ok := d.GetOk("destination_file"); ok {
			f.destinationFile = v.(string)
		} else {
			return fmt.Errorf("destination_file argument is required")
		}

		client := meta.(*govmomi.Client)
		dc, err := getDatacenter(client, f.datacenter)
		if err != nil {
			return err
		}

		finder := find.NewFinder(client.Client, true)
		finder = finder.SetDatacenter(dc)

		ds, err := getDatastore(finder, f.datastore)
		if err != nil {
			return fmt.Errorf("error %s", err)
		}

		fm := object.NewFileManager(client.Client)
		task, err := fm.MoveDatastoreFile(context.TODO(), ds.Path(oldDestinationFile.(string)), dc, ds.Path(newDestinationFile.(string)), dc, true)
		if err != nil {
			return err
		}

		_, err = task.WaitForResult(context.TODO(), nil)
		if err != nil {
			return err
		}

	}

	return nil
}

func resourceVSphereDvsDelete(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] deleting file: %#v", d)
	f := file{}

	if v, ok := d.GetOk("datacenter"); ok {
		f.datacenter = v.(string)
	}

	if v, ok := d.GetOk("datastore"); ok {
		f.datastore = v.(string)
	} else {
		return fmt.Errorf("datastore argument is required")
	}

	if v, ok := d.GetOk("source_file"); ok {
		f.sourceFile = v.(string)
	} else {
		return fmt.Errorf("source_file argument is required")
	}

	if v, ok := d.GetOk("destination_file"); ok {
		f.destinationFile = v.(string)
	} else {
		return fmt.Errorf("destination_file argument is required")
	}

	client := meta.(*govmomi.Client)

	err := deleteFile(client, &f)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

