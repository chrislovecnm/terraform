package vsphere

import (
	"fmt"
	"log"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type StoragePodDataStore struct {
	name           string
	template       string
	storagePodName string

	ConfigSpecsNetwork []types.BaseVirtualDeviceConfigSpec
	ResourcePool       *object.ResourcePool
	// TODO what is this??
	// HostSystem         *object.HostSystem
	Folder         *object.Folder
	VirtualMachine *object.VirtualMachine
	DataCenter     *object.Datacenter
}

// Based of of govc clone.go
// Get the recommended StoragePod datastore
func (spds *StoragePodDataStore) findRecommenedStoragePodDataStore(client *vim25.Client) (datastore *object.Datastore, err error) {
	datastoreref := types.ManagedObjectReference{}

	folderref := spds.Folder.Reference()
	poolref := spds.ResourcePool.Reference()

	relocateSpec := types.VirtualMachineRelocateSpec{
		DeviceChange: spds.ConfigSpecsNetwork,
		Folder:       &folderref,
		Pool:         &poolref,
	}

	// TODO what is this?
	// if pds.HostSystem != nil {
	// 	hostref := pds.HostSystem.Reference()
	// 	relocateSpec.Host = &hostref
	// }

	hasTemplate := false
	if spds.template != "" {
		hasTemplate = true
	}
	cloneSpec := &types.VirtualMachineCloneSpec{
		Location: relocateSpec,
		PowerOn:  false,
		Template: hasTemplate,
	}

	sp, err := spds.findStoragePod(client)
	if err != nil {
		log.Printf("[ERROR] Couldn't find storage pod '%s'.  %s", spds.storagePodName, err)
		return nil, err
	}
	storagePod := sp.Reference()

	// Build pod selection spec from config spec
	podSelectionSpec := types.StorageDrsPodSelectionSpec{
		StoragePod: &storagePod,
	}

	// Get the virtual machine reference
	vmref := spds.VirtualMachine.Reference()

	// Build the placement spec
	storagePlacementSpec := types.StoragePlacementSpec{
		Folder:           &folderref,
		Vm:               &vmref,
		CloneName:        spds.name,
		CloneSpec:        cloneSpec,
		PodSelectionSpec: podSelectionSpec,
		Type:             string(types.StoragePlacementSpecPlacementTypeClone),
	}
	log.Printf("[DEBUG] storage placement spec, %v", storagePlacementSpec)

	// Get the storage placement result
	storageResourceManager := object.NewStorageResourceManager(client)
	var result *types.StoragePlacementResult
	result, err = storageResourceManager.RecommendDatastores(context.TODO(), storagePlacementSpec)
	if err != nil {
		log.Printf("[ERROR] Couldn't find datastore cluster %v.  %s", spds.storagePodName, err)
		return nil, err
	}

	// Get the recommendations
	recommendations := result.Recommendations
	if len(recommendations) == 0 {
		log.Printf("[ERROR] no recommendations for datastore")
		return nil, fmt.Errorf("no recommendations for datastore")
	}

	// Get the first recommendation
	datastoreref = recommendations[0].Action[0].(*types.StoragePlacementAction).Destination
	datastore = object.NewDatastore(client, datastoreref)
	log.Printf("[DEBUG] Found datastore: %v", datastore)
	return datastore, nil
}

// find the Default or named Storage Pod
func (spds *StoragePodDataStore) findStoragePod(client *vim25.Client) (sp *object.StoragePod, err error) {

	finder := find.NewFinder(client, true)
	if spds.DataCenter != nil {
		finder.SetDatacenter(spds.DataCenter)
	}

	if spds.storagePodName != "" {
		log.Printf("[DEBUG] looking for DataStore Cluster")
		sp, err = finder.DatastoreCluster(context.TODO(), spds.storagePodName)
	} else {
		// TODO this does not seem to be working ... wth
		sp, err = finder.DefaultDatastoreCluster(context.TODO())
	}

	if err != nil {
		log.Printf("[ERROR] Couldn't find datastore cluster %v.  %s", spds.storagePodName, err)
		return nil, err
	}

	log.Printf("[DEBUG] Found datastore cluster: %v", sp)
	return sp, nil
}
