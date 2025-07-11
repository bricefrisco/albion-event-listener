package main

import (
	"errors"
	"golang.org/x/sys/windows"
	"os"
	"syscall"
	"unsafe"
)

func adapterAddresses() ([]*windows.IpAdapterAddresses, error) {
	var b []byte
	l := uint32(15000) // recommended initial size
	for {
		b = make([]byte, l)
		err := windows.GetAdaptersAddresses(syscall.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_PREFIX, 0, (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])), &l)
		if err == nil {
			if l == 0 {
				return nil, nil
			}
			break
		}
		if !errors.Is(err.(syscall.Errno), syscall.ERROR_BUFFER_OVERFLOW) {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
		if l <= uint32(len(b)) {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
	}
	var aas []*windows.IpAdapterAddresses
	for aa := (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])); aa != nil; aa = aa.Next {
		aas = append(aas, aa)
	}
	return aas, nil
}

func bytePtrToString(p *uint8) string {
	a := (*[10000]uint8)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return string(a[:i])
}

func bytePtrToString16(p *uint16) string {
	a := (*[10000]uint16)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return syscall.UTF16ToString(a[:i])
}

func getPhysicalInterfaces() ([]string, error) {
	addresses, err := adapterAddresses()
	if err != nil {
		return nil, err
	}

	var result []string
	for _, address := range addresses {
		friendlyName := bytePtrToString16(address.FriendlyName)
		if friendlyName == "NordLynx" || friendlyName == "Wi-Fi" {
			result = append(result, "\\Device\\NPF_"+bytePtrToString(address.AdapterName))
		}
	}

	if len(result) == 0 {
		return nil, errors.New("could not find a physical interface")
	}

	return result, nil
}
