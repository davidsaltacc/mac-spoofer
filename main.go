package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type AdapterInfo struct {
	Value string
	Index int
}

func getKey(id int) registry.Key {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, "SYSTEM\\CurrentControlSet\\Control\\Class\\{4d36e972-e325-11ce-bfc1-08002be10318}\\"+fmt.Sprintf("%04d", id), registry.ALL_ACCESS)
	if err != nil {
		panic(err)
	}
	return key
}

func changeAdapter(id int, disable bool) {
	action := "enable"
	if disable {
		action = "disable"
	}
	cmd := exec.Command("netsh", "interface", "set", "interface", "\""+getNetworkAdapterInfo("NetConnectionId")[id]+"\"", "admin="+action)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err := cmd.Output()
	if err != nil {
		panic(err)
	}
}

func removeColons(mac string) string {
	new := strings.ReplaceAll(mac, ":", "")
	if len(new) != 12 {
		panic("invalid mac address supplied")
	}
	return new
}

func addColons(mac string) string {
	if len(mac) != 12 {
		panic("invalid mac address supplied")
	}
	digits := strings.Split(mac, "")
	newMac := ""
	for i := 0; i < 12; i += 2 {
		newMac = newMac + fmt.Sprintf("%s%s:", digits[i], digits[i+1])
	}
	return newMac[:len(newMac)-1]
}

func randomMac() string {
	return strings.ToUpper(fmt.Sprintf(
		"%02x:%02x:%02x:%02x:%02x:%02x",
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
	))
}

func getCurrentMac(index int) string {
	key := getKey(index)
	value, _, err := key.GetStringValue("NetworkAddress")
	if err != nil {
		return getNetworkAdapterInfo("MACAddress")[index]
	}
	return addColons(value)
}

func getOriginalMac(index int) string {
	key := getKey(index)
	value, _, err := key.GetStringValue("NetworkAddressOrig")
	if err != nil {
		return getCurrentMac(index)
	}
	return addColons(value)
}

func isValidMac(mac string) bool {
	matched, _ := regexp.MatchString(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`, mac)
	return matched
}

func setMac(mac string, index int) {
	key := getKey(index)
	oldMac := getOriginalMac(index)
	err := key.SetStringValue("NetworkAddressOrig", removeColons(oldMac))
	if err != nil {
		panic(err)
	}
	changeAdapter(index, true)
	err = key.SetStringValue("NetworkAddress", removeColons(mac))
	changeAdapter(index, false)
	if err != nil {
		panic(err)
	}
}

func getNetworkAdapterInfo(query string) []string {
	cmd := exec.Command("wmic", "nic", "get", query, "/format:csv")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	data, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	adaptersRaw := strings.ReplaceAll(string(data), "\r", "")
	adapters := strings.Split(adaptersRaw, "\n")
	for i, adapter := range adapters {
		adapters[i] = strings.Join(strings.Split(adapter, ",")[1:], "")
	}
	return adapters[2:]
}

func main() {

	var mw *walk.MainWindow
	var macInput *walk.LineEdit
	var statusLabel *walk.Label
	var adapterCombo *walk.ComboBox

	rawAdapters := getNetworkAdapterInfo("NetConnectionId")
	var validAdapters []AdapterInfo

	for i, name := range rawAdapters {
		if strings.TrimSpace(name) != "" {
			validAdapters = append(validAdapters, AdapterInfo{
				Value: name,
				Index: i,
			})
		}
	}

	adapters := []AdapterInfo{}
	adapters = append(adapters, validAdapters...)

	if _, err := (MainWindow{
		AssignTo: &mw,
		Title:    "MAC Spoofer",
		Size:     Size{300, 200},
		Layout:   VBox{},
		Children: []Widget{
			Label{
				Text: "Select Network Adapter:",
			},
			ComboBox{
				AssignTo:      &adapterCombo,
				Model:         adapters,
				DisplayMember: "Value",
				OnCurrentIndexChanged: func() {
					if adapterCombo.CurrentIndex() >= 0 && adapterCombo.Focused() {
						macInput.SetText(getCurrentMac(adapters[adapterCombo.CurrentIndex()].Index))
					}
				},
				CurrentIndex: -1,
			},
			Label{
				Text: "Enter MAC Address:",
			},
			LineEdit{
				AssignTo:    &macInput,
				ToolTipText: "Enter MAC address in XX:XX:XX:XX:XX:XX format",
				Text:        "",
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					PushButton{
						Text: "Randomize",
						OnClicked: func() {
							if adapterCombo.CurrentIndex() < 0 {
								statusLabel.SetText("Please select an adapter")
								return
							}
							random := randomMac()
							setMac(random, adapters[adapterCombo.CurrentIndex()].Index)
							macInput.SetText(random)
							statusLabel.SetText("Randomized MAC applied")
						},
					},
					PushButton{
						Text: "Restore Original",
						OnClicked: func() {
							if adapterCombo.CurrentIndex() < 0 {
								statusLabel.SetText("Please select an adapter")
								return
							}
							original := getOriginalMac(adapters[adapterCombo.CurrentIndex()].Index)
							setMac(original, adapters[adapterCombo.CurrentIndex()].Index)
							macInput.SetText(original)
							statusLabel.SetText("Original MAC restored")
						},
					},
				},
			},
			PushButton{
				Text: "Apply MAC Address",
				Font: Font{Bold: true},
				OnClicked: func() {
					if adapterCombo.CurrentIndex() < 0 {
						statusLabel.SetText("Please select an adapter")
						return
					}
					macInput.SetText(strings.ToUpper(macInput.Text()))
					currentMAC := macInput.Text()
					if isValidMac(currentMAC) {
						println("attempting to set mac address to " + currentMAC)
						setMac(currentMAC, adapters[adapterCombo.CurrentIndex()].Index)
						statusLabel.SetText("MAC address applied successfully")
					} else {
						setMac(currentMAC, adapters[adapterCombo.CurrentIndex()].Index)
						statusLabel.SetText("Invalid MAC address format")
					}
				},
			},
			Label{
				AssignTo: &statusLabel,
				Text:     "Ready",
			},
		},
	}).Run(); err != nil {
		panic(err)
	}
}
