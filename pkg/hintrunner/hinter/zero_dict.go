package hinter

import (
	"fmt"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// Used to keep track of all dictionaries data
type ZeroDictionary struct {
	// The Data contained in a dictionary
	Data map[f.Element]mem.MemoryValue
	// Default value for key not present in the dictionary
	DefaultValue mem.MemoryValue
	// first free offset in memory segment of dictionary
	FreeOffset *uint64
}

// Gets the memory value at certain key
func (d *ZeroDictionary) at(key f.Element) (mem.MemoryValue, error) {
	if value, ok := d.Data[key]; ok {
		return value, nil
	}
	if d.DefaultValue != mem.UnknownValue {
		return d.DefaultValue, nil
	}
	return mem.UnknownValue, fmt.Errorf("no value for key: %v", key)
}

// Given a key and a value, it sets the value at the given key
func (d *ZeroDictionary) set(key f.Element, value mem.MemoryValue) {
	d.Data[key] = value
}

// Given a incrementBy value, it increments the freeOffset field of dictionary by it
func (d *ZeroDictionary) incrementFreeOffset(freeOffset uint64) {
	*d.FreeOffset += freeOffset
}

// Given a freeOffset value, it sets the freeOffset field of dictionary to it
func (d *ZeroDictionary) setFreeOffset(freeOffset uint64) {
	*d.FreeOffset = freeOffset
}

// Used to manage dictionaries creation
type ZeroDictionaryManager struct {
	// a map that links a segment index to a dictionary
	Dictionaries map[uint64]ZeroDictionary
}

func NewZeroDictionaryManager() ZeroDictionaryManager {
	return ZeroDictionaryManager{
		Dictionaries: make(map[uint64]ZeroDictionary),
	}
}

// It creates a new segment which will hold dictionary values. It links this
// segment with the current dictionary and returns the address that points
// to the start of this segment. initial dictionary data is set from the data argument.
func (dm *ZeroDictionaryManager) NewDictionary(vm *VM.VirtualMachine, data map[f.Element]mem.MemoryValue) mem.MemoryAddress {
	newDictAddr := vm.Memory.AllocateEmptySegment()
	freeOffset := uint64(0)
	dm.Dictionaries[newDictAddr.SegmentIndex] = ZeroDictionary{
		Data:         data,
		DefaultValue: mem.UnknownValue,
		FreeOffset:   &freeOffset,
	}
	return newDictAddr
}

// It creates a new segment which will hold dictionary values. It links this
// segment with the current dictionary and returns the address that points
// to the start of this segment. If key not present in the dictionary during
// querying the defaultValue will be returned instead.
func (dm *ZeroDictionaryManager) NewDefaultDictionary(vm *VM.VirtualMachine, defaultValue mem.MemoryValue) mem.MemoryAddress {
	newDefaultDictAddr := vm.Memory.AllocateEmptySegment()
	freeOffset := uint64(0)
	dm.Dictionaries[newDefaultDictAddr.SegmentIndex] = ZeroDictionary{
		Data:         make(map[f.Element]mem.MemoryValue),
		DefaultValue: defaultValue,
		FreeOffset:   &freeOffset,
	}
	return newDefaultDictAddr
}

// Given a memory address, it looks for the right dictionary using the segment index. If no
// segment is associated with the given segment index, it errors
func (dm *ZeroDictionaryManager) GetDictionary(dictAddr mem.MemoryAddress) (ZeroDictionary, error) {
	dict, ok := dm.Dictionaries[dictAddr.SegmentIndex]
	if ok {
		return dict, nil
	}
	return ZeroDictionary{}, fmt.Errorf("no dictionary at address: %s", dictAddr)
}

// Given a memory address and a key it returns the value held at that position. The address is used
// to locate the correct dictionary and the key to index on it
func (dm *ZeroDictionaryManager) At(dictAddr mem.MemoryAddress, key f.Element) (mem.MemoryValue, error) {
	if dict, ok := dm.dictionaries[dictAddr.SegmentIndex]; ok {
		return dict.at(key)
	}
	return mem.UnknownValue, fmt.Errorf("no dictionary at address: %s", dictAddr)
}

// Given a memory address,a key and a value it stores the value at the correct position.
func (dm *ZeroDictionaryManager) Set(dictAddr mem.MemoryAddress, key f.Element, value mem.MemoryValue) error {
	if dict, ok := dm.dictionaries[dictAddr.SegmentIndex]; ok {
		dict.set(key, value)
		return nil
	}
	return fmt.Errorf("no dictionary at address: %s", dictAddr)
}

// Given a memory address and a incrementBy, it increments the freeOffset field of dictionary by it.
func (dm *ZeroDictionaryManager) IncrementFreeOffset(dictAddr mem.MemoryAddress, incrementBy uint64) error {
	if dict, ok := dm.dictionaries[dictAddr.SegmentIndex]; ok {
		dict.incrementFreeOffset(incrementBy)
		return nil
	}
	return fmt.Errorf("no dictionary at address: %s", dictAddr)
}

// Given a memory address and a freeOffset, it sets the freeOffset field of dictionary to it.
func (dm *ZeroDictionaryManager) SetFreeOffset(dictAddr mem.MemoryAddress, freeOffset uint64) error {
	if dict, ok := dm.dictionaries[dictAddr.SegmentIndex]; ok {
		dict.setFreeOffset(freeOffset)
		return nil
	}
	return fmt.Errorf("no dictionary at address: %s", dictAddr)
}

// CopyZeroDictionary creates a copy of a ZeroDictionary
func CopyZeroDictionary(dict *ZeroDictionary) (*ZeroDictionary, error) {
	// Copy the Data field
	dataCopy := make(map[f.Element]mem.MemoryValue)
	for k, v := range dict.Data {
		// Copy the key
		keyCopy := f.Element{}
		keyCopy.Set(&k)

		// Copy the value
		feltCopy := f.Element{}
		feltCopy.Set(&v.Felt)

		valueCopy := mem.MemoryValue{
			Felt: feltCopy,
			Kind: v.Kind,
		}

		dataCopy[keyCopy] = valueCopy
	}

	// Copy the DefaultValue field
	defaultValueCopy := dict.DefaultValue

	// Copy the FreeOffset field
	var freeOffsetCopy *uint64
	if dict.FreeOffset != nil {
		offsetValue := *dict.FreeOffset
		freeOffsetCopy = &offsetValue
	}

	// Create and return the new instance of ZeroDictionary
	return &ZeroDictionary{
		Data:         dataCopy,
		DefaultValue: defaultValueCopy,
		FreeOffset:   freeOffsetCopy,
	}, nil
}
