package vm

import (
	"log"
	vbox "github.com/rmxymh/go-virtualbox"
)

const (
	BOOT_DEVICE_PXE =	"net"
	BOOT_DEVICE_DISK =	"disk"
	BOOT_DEVICE_CD_DVD =	"dvd"
	BOOT_DEVICE_FLOPPY =	"floppy"
)

type Instance struct {
	Name string
	FakeNode		bool

	lastBootOrder		[]string
	nextBootOrder		[]string
	needToRestoreBootOrder	bool
	changeBootOrder		bool
}

var instances map[string]Instance

func init() {
	instances = make(map[string]Instance)
}

func AddInstnace(name string, fakeNode bool) Instance {
	newInstance := Instance {
		Name: name,
		FakeNode: fakeNode,
	}
	instances[name] = newInstance
	log.Println("Add instance ", name)

	return newInstance
}

func DeleteInstance(name string) {
	_, ok := instances[name]
	if ok {
		delete(instances, name)
	}
	log.Println("Remove instance ", name)
}

func GetInstance(name string) (instance Instance, ok bool) {
	instance, ok = instances[name]
	return instance, ok
}

func (instance *Instance)IsRunning() bool {
	if instance.FakeNode {
		return true
	}

	machine, err := vbox.GetMachine(instance.Name)

	if err == nil && machine.State == vbox.Running {
		return true
	}
	return false
}

func (instance *Instance)SetBootDevice(dev string) {
	if instance.FakeNode {
		return
	}

	machine, err := vbox.GetMachine(instance.Name)

	if err != nil {
		log.Fatalf("    Instance: Failed to set BootDevice to VM %s: %s", instance.Name, err.Error())
		return
	}

	newBootOrder := []string{dev}
	for _, d := range machine.BootOrder {
		if d != dev {
			newBootOrder = append(newBootOrder, d)
		}
	}

	instance.nextBootOrder = newBootOrder
	instance.changeBootOrder = true
}

func (instance *Instance)PowerOff() {
	if instance.FakeNode {
		return
	}

	machine, err := vbox.GetMachine(instance.Name)

	if err != nil {
		log.Fatalf("    Instance: Failed to find VM %s and power off it: %s", instance.Name, err.Error())
		return
	}

	machine.Poweroff()

	if instance.needToRestoreBootOrder {
		machine.BootOrder = instance.lastBootOrder
		machine.Modify()
		instance.lastBootOrder = make([]string, 4)
		instance.needToRestoreBootOrder = false
	}
}

func (instance *Instance)ACPIOff() {
	if instance.FakeNode {
		return
	}

	machine, err := vbox.GetMachine(instance.Name)

	if err != nil {
		log.Fatalf("    Instance: Failed to find VM %s and power off it: %s", instance.Name, err.Error())
		return
	}

	machine.Stop()

	if instance.needToRestoreBootOrder {
		machine.BootOrder = instance.lastBootOrder
		machine.Modify()
		instance.lastBootOrder = make([]string, 4)
		instance.needToRestoreBootOrder = false
	}
}

func (instance *Instance)PowerOn() {
	if instance.FakeNode {
		return
	}

	machine, err := vbox.GetMachine(instance.Name)

	if err != nil {
		log.Fatalf("    Instance: Failed to find VM %s and power on it: %s", instance.Name, err.Error())
		return
	}

	if instance.changeBootOrder {
		instance.lastBootOrder = machine.BootOrder
		machine.BootOrder = instance.nextBootOrder
		machine.Modify()
		instance.nextBootOrder = make([]string, 4)
		instance.changeBootOrder = false
		log.Println("Current Boot Order = ", machine.BootOrder)
		instance.needToRestoreBootOrder = true
	}

	machine.Start()
}

func (instance *Instance)Reset() {
	if instance.FakeNode {
		return
	}

	machine, err := vbox.GetMachine(instance.Name)

	if err != nil {
		log.Fatalf("    Instance: Failed to find VM %s and power on it: %s", instance.Name, err.Error())
		return
	}

	machine.Reset()
}