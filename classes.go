package quickjs

import "sync"

type ClassId uint32

var classRefs struct {
	sync.RWMutex
	nameToId  map[string]ClassId
	idToClass map[ClassId]*Class
}

func init() {
	classRefs.Lock()
	defer classRefs.Unlock()
	classRefs.nameToId = make(map[string]ClassId)
	classRefs.idToClass = make(map[ClassId]*Class)
}

func AddClassId(name string, id ClassId) {
	classRefs.Lock()
	defer classRefs.Unlock()

	classRefs.nameToId[name] = id
}

func AddClassObject(id ClassId, class *Class) {
	classRefs.Lock()
	defer classRefs.Unlock()
	classRefs.idToClass[id] = class
}

func (id ClassId) Get() (*Class, bool) {
	classRefs.RLock()
	defer classRefs.RUnlock()
	obj, ok := classRefs.idToClass[id]
	return obj, ok
}

func FindClassId(name string) (ClassId, bool) {
	classRefs.RLock()
	defer classRefs.RUnlock()

	obj, ok := classRefs.nameToId[name]
	return obj, ok
}
