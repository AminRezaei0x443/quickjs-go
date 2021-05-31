package quickjs

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"unsafe"
)

/*
#cgo CFLAGS: -I./3rdparty/include/quickjs
#cgo linux,!android,386 LDFLAGS: -L${SRCDIR}/3rdparty/libs/quickjs/linux/x86 -lquickjs
#cgo linux,!android,amd64 LDFLAGS: -L${SRCDIR}/3rdparty/libs/quickjs/Linux -lquickjs
//#cgo linux,!android,amd64 LDFLAGS: -L${SRCDIR}/3rdparty/libs/quickjs/linux/x86_64 -lquickjs
#cgo linux,!android LDFLAGS: -lm -ldl -lpthread
#cgo windows,386 LDFLAGS: -L${SRCDIR}/3rdparty/libs/quickjs/windows/x86 -lquickjs
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/3rdparty/libs/quickjs/windows/x86_64 -lquickjs
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/3rdparty/libs/quickjs/darwin/x86_64 -lquickjs
#cgo android,386 LDFLAGS: -L${SRCDIR}/3rdparty/libs/quickjs/Android/x86 -lquickjs
#cgo android,amd64 LDFLAGS: -L${SRCDIR}/3rdparty/libs/quickjs/Android/x86_64 -lquickjs
#cgo android,arm LDFLAGS: -L${SRCDIR}/3rdparty/libs/quickjs/Android/armeabi-v7a -lquickjs
#cgo android,arm64 LDFLAGS: -L${SRCDIR}/3rdparty/libs/quickjs/Android/arm64-v8a -lquickjs
#cgo android LDFLAGS: -landroid -llog -lm

#include <stdlib.h>
#include <string.h>
#include "quickjs.h"
#include "quickjs-libc.h"

extern JSValue proxy(JSContext *ctx, JSValueConst this_val, int argc, JSValueConst *argv);
extern int moduleInitProxy(JSContext *ctx, JSModuleDef* mod);
extern void finalizeObjectInstance(JSRuntime* ctx, JSValue value);

static JSValue JS_NewNull() { return JS_NULL; }
static JSValue JS_NewUndefined() { return JS_UNDEFINED; }
static JSValue JS_NewUninitialized() { return JS_UNINITIALIZED; }

static JSValue ThrowSyntaxError(JSContext *ctx, const char *fmt) { return JS_ThrowSyntaxError(ctx, "%s", fmt); }
static JSValue ThrowTypeError(JSContext *ctx, const char *fmt) { return JS_ThrowTypeError(ctx, "%s", fmt); }
static JSValue ThrowReferenceError(JSContext *ctx, const char *fmt) { return JS_ThrowReferenceError(ctx, "%s", fmt); }
static JSValue ThrowRangeError(JSContext *ctx, const char *fmt) { return JS_ThrowRangeError(ctx, "%s", fmt); }
static JSValue ThrowInternalError(JSContext *ctx, const char *fmt) { return JS_ThrowInternalError(ctx, "%s", fmt); }

int has_suffix(const char *str, const char *suffix);

static int eval_buf(JSContext *ctx, const void *buf, int buf_len,
                    const char *filename, int eval_flags)
{
    JSValue val;
    int ret;

    if ((eval_flags & JS_EVAL_TYPE_MASK) == JS_EVAL_TYPE_MODULE) {
        //for the modules, we compile then run to be able to set import.meta
        val = JS_Eval(ctx, buf, buf_len, filename,
                      eval_flags | JS_EVAL_FLAG_COMPILE_ONLY);
        if (!JS_IsException(val)) {
            js_module_set_import_meta(ctx, val, 1, 1);
            val = JS_EvalFunction(ctx, val);
        }
    } else {
        val = JS_Eval(ctx, buf, buf_len, filename, eval_flags);
    }
    if (JS_IsException(val)) {
        js_std_dump_error(ctx);
        ret = -1;
    } else {
        ret = 0;
    }

    JS_FreeValue(ctx, val);
    return ret;
}

static int eval_file(JSContext *ctx, const char *filename, int module)
{
    uint8_t *buf;
    int ret, eval_flags;
    size_t buf_len;

    buf = js_load_file(ctx, &buf_len, filename);
    if (!buf) {
        perror(filename);
        exit(1);
    }

    if (module < 0) {
        module = (has_suffix(filename, ".mjs") ||
                  JS_DetectModule((const char *)buf, buf_len));
    }
    if (module)
        eval_flags = JS_EVAL_TYPE_MODULE;
    else
        eval_flags = JS_EVAL_TYPE_GLOBAL;
    ret = eval_buf(ctx, buf, buf_len, filename, eval_flags);
    js_free(ctx, buf);
    return ret;
}

static int SetBaseGlobal(JSContext *ctx) {
	const char *str = "import * as std from 'std';\n"
                "import * as os from 'os';\n"
                "globalThis.std = std;\n"
                "globalThis.os = os;\n";
    return eval_buf(ctx, str, strlen(str), "<input>", JS_EVAL_TYPE_MODULE);
}

static JSContext *JS_NewCustomContext(JSRuntime *rt)
{
    JSContext *ctx;
    ctx = JS_NewContext(rt);
    if (!ctx)
        return NULL;

	JS_AddIntrinsicBigFloat(ctx);
	JS_AddIntrinsicBigDecimal(ctx);
	JS_AddIntrinsicOperators(ctx);
	JS_EnableBignumExt(ctx, 1);

    js_init_module_std(ctx, "std");
    js_init_module_os(ctx, "os");
    return ctx;
}

static JSRuntime* NewJsRuntime() {
	JSRuntime *rt = JS_NewRuntime();
	js_std_set_worker_new_context_func(JS_NewCustomContext);
    js_std_init_handlers(rt);

	return rt;
}

static JSContext* NewJsContext(JSRuntime *rt) {
	js_std_set_worker_new_context_func(JS_NewCustomContext);
    js_std_init_handlers(rt);
    JSContext* ctx = JS_NewCustomContext(rt);

	// loader for ES6 modules
    JS_SetModuleLoaderFunc(rt, NULL, js_module_loader, NULL);

	js_std_add_helpers(ctx, -1, NULL);
	SetBaseGlobal(ctx);
	js_std_loop(ctx);

	return ctx;
}

static JSClassDef* NewClassDef(const char* name){
	JSClassDef* d = malloc(sizeof(JSClassDef));
	d->class_name = name;
	d->finalizer = finalizeObjectInstance;
	return d;
}
extern void freeArrayBuffer(JSRuntime *rt, void *opaque, void *ptr);
*/
import "C"

var (
	EVAL_GLOBAL int = int(C.JS_EVAL_TYPE_GLOBAL)
	EVAL_MODULE int = int(C.JS_EVAL_TYPE_MODULE)
	EVAL_STRICT int = int(C.JS_EVAL_FLAG_STRICT)
	EVAL_STRIP  int = int(C.JS_EVAL_FLAG_STRIP)
)

type Runtime struct {
	ref *C.JSRuntime
}

func NewRuntime() Runtime {
	rt := Runtime{ref: C.NewJsRuntime()}
	C.JS_SetCanBlock(rt.ref, C.int(1))
	return rt
}

func (r Runtime) RunGC() { C.JS_RunGC(r.ref) }

func (r Runtime) Free() {
	C.JS_FreeRuntime(r.ref)
}

func (r Runtime) StdFreeHandlers() {
	C.js_std_free_handlers(r.ref)
}

func (r Runtime) NewContext() *Context {
	ref := C.NewJsContext(r.ref)

	return &Context{ref: ref}
}

func WrapContext(ref unsafe.Pointer) *Context {
	cref := (*C.JSContext)(ref)
	return &Context{ref: cref}
}
func WrapRuntime(ref unsafe.Pointer) *Runtime {
	cref := (*C.JSRuntime)(ref)
	return &Runtime{ref: cref}
}

func (r Runtime) ExecutePendingJob() (Context, error) {
	var ctx Context

	err := C.JS_ExecutePendingJob(r.ref, &ctx.ref)
	if err <= 0 {
		if err == 0 {
			return ctx, io.EOF
		}
		return ctx, ctx.Exception()
	}

	return ctx, nil
}

type Function func(ctx *Context, this Value, args []Value) Value

type funcEntry struct {
	ctx *Context
	fn  Function
}

func storeFuncPtr(v *funcEntry) ObjectId {
	return NewObjectId(v)
}

func restoreFuncPtr(id ObjectId) *funcEntry {
	if v, ok := id.Get(); ok {
		if _v, ok := v.(*funcEntry); ok {
			return _v
		}
	}

	return nil
}

//export proxy
func proxy(ctx *C.JSContext, thisVal C.JSValueConst, argc C.int, argv *C.JSValueConst) C.JSValue {
	refs := (*[1 << unsafe.Sizeof(0)]C.JSValueConst)(unsafe.Pointer(argv))[:argc:argc]

	id := C.int64_t(0)
	C.JS_ToInt64(ctx, &id, refs[0])

	entry := restoreFuncPtr(ObjectId(id))

	args := make([]Value, len(refs)-1)
	for i := 0; i < len(args); i++ {
		args[i].ctx = entry.ctx
		args[i].ref = refs[1+i]
	}

	result := entry.fn(entry.ctx, Value{ctx: entry.ctx, ref: thisVal}, args)

	return result.ref
}

type Context struct {
	ref       *C.JSContext
	globals   *Value
	proxy     *Value
	ctorProxy *Value
}

func (ctx *Context) Free() {
	if ctx.proxy != nil {
		ctx.proxy.Free()
	}
	if ctx.ctorProxy != nil {
		ctx.ctorProxy.Free()
	}
	if ctx.globals != nil {
		ctx.globals.Free()
	}

	C.JS_FreeContext(ctx.ref)
}

func (ctx *Context) FreeVals() {
	if ctx.proxy != nil {
		ctx.proxy.Free()
	}
	if ctx.ctorProxy != nil {
		ctx.ctorProxy.Free()
	}
	if ctx.globals != nil {
		ctx.globals.Free()
	}
}

func (ctx *Context) Function(fn Function) Value {
	val := ctx.eval(`(proxy, id) => function() { return proxy.call(this, id, ...arguments); }`)
	if val.IsException() {
		return val
	}
	defer val.Free()

	funcPtr := storeFuncPtr(&funcEntry{ctx: ctx, fn: fn})
	funcPtrVal := ctx.Int64(int64(funcPtr))

	if ctx.proxy == nil {
		ctx.proxy = &Value{
			ctx: ctx,
			ref: C.JS_NewCFunction(ctx.ref, (*C.JSCFunction)(unsafe.Pointer(C.proxy)), nil, C.int(0)),
		}
	}

	args := []C.JSValue{ctx.proxy.ref, funcPtrVal.ref}

	return Value{ctx: ctx, ref: C.JS_Call(ctx.ref, val.ref, ctx.Null().ref, C.int(len(args)), &args[0])}
}

func (ctx *Context) DupValue(value Value) Value {
	return Value{ctx: ctx, ref: C.JS_DupValue(ctx.ref, value.ref)}
}

func (ctx *Context) JsFunction(this Value, fn Value, args []Value) Value {
	if fn.IsFunction() {
		var _argsptr *C.JSValue = nil
		if len(args) != 0 {
			_args := make([]C.JSValue, len(args))
			for i := 0; i < len(args); i++ {
				_args[i] = args[i].ref
			}

			_argsptr = &_args[0]
		}

		return Value{ctx: ctx, ref: C.JS_Call(ctx.ref, fn.ref, this.ref, C.int(len(args)), _argsptr)}
	} else {
		return ctx.Error(errors.New("fn must be function"))
	}
}

func (ctx *Context) Null() Value {
	return Value{ctx: ctx, ref: C.JS_NewNull()}
}

func (ctx *Context) Undefined() Value {
	return Value{ctx: ctx, ref: C.JS_NewUndefined()}
}

func (ctx *Context) Uninitialized() Value {
	return Value{ctx: ctx, ref: C.JS_NewUninitialized()}
}

func (ctx *Context) Error(err error) Value {
	val := Value{ctx: ctx, ref: C.JS_NewError(ctx.ref)}
	val.Set("message", ctx.String(err.Error()))
	return val
}

func (ctx *Context) Bool(b bool) Value {
	bv := 0
	if b {
		bv = 1
	}
	return Value{ctx: ctx, ref: C.JS_NewBool(ctx.ref, C.int(bv))}
}

func (ctx *Context) Int32(v int32) Value {
	return Value{ctx: ctx, ref: C.JS_NewInt32(ctx.ref, C.int32_t(v))}
}

func (ctx *Context) Int64(v int64) Value {
	return Value{ctx: ctx, ref: C.JS_NewInt64(ctx.ref, C.int64_t(v))}
}

func (ctx *Context) Uint32(v uint32) Value {
	return Value{ctx: ctx, ref: C.JS_NewUint32(ctx.ref, C.uint32_t(v))}
}

func (ctx *Context) BigInt64(v uint64) Value {
	return Value{ctx: ctx, ref: C.JS_NewBigInt64(ctx.ref, C.int64_t(v))}
}

func (ctx *Context) BigUint64(v uint64) Value {
	return Value{ctx: ctx, ref: C.JS_NewBigUint64(ctx.ref, C.uint64_t(v))}
}

func (ctx *Context) Float64(v float64) Value {
	return Value{ctx: ctx, ref: C.JS_NewFloat64(ctx.ref, C.double(v))}
}

func (ctx *Context) String(v string) Value {
	ptr := C.CString(v)
	defer C.free(unsafe.Pointer(ptr))
	return Value{ctx: ctx, ref: C.JS_NewString(ctx.ref, ptr)}
}

func (ctx *Context) Atom(v string) Atom {
	ptr := C.CString(v)
	defer C.free(unsafe.Pointer(ptr))
	return Atom{ctx: ctx, ref: C.JS_NewAtom(ctx.ref, ptr)}
}

func (ctx *Context) eval(code string) Value { return ctx.evalFile(code, 0, "<code>") }

func (ctx *Context) evalFile(code string, evaltype int, filename string) Value {
	codePtr := C.CString(code)
	defer C.free(unsafe.Pointer(codePtr))

	filenamePtr := C.CString(filename)
	defer C.free(unsafe.Pointer(filenamePtr))

	return Value{ctx: ctx, ref: C.JS_Eval(ctx.ref, codePtr, C.size_t(len(code)), filenamePtr, C.int(evaltype))}
}

func (ctx *Context) Eval(code string, evaltype int) (Value, error) {
	return ctx.EvalFile(code, evaltype, "<code>")
}

func (ctx *Context) EvalFile(code string, evaltype int, filename string) (Value, error) {
	val := ctx.evalFile(code, evaltype, filename)

	if val.IsException() {
		return val, ctx.Exception()
	}
	return val, nil
}

func (ctx *Context) Call(this Value, fn Value, args []Value) (Value, error) {
	val := ctx.JsFunction(this, fn, args)
	if val.IsException() {
		err := ctx.Exception()
		return val, err
	} else {
		return val, nil
	}
}

func (ctx *Context) Globals() Value {
	if ctx.globals == nil {
		ctx.globals = &Value{
			ctx: ctx,
			ref: C.JS_GetGlobalObject(ctx.ref),
		}
	}
	return *ctx.globals
}

func (ctx *Context) Throw(v Value) Value {
	return Value{ctx: ctx, ref: C.JS_Throw(ctx.ref, v.ref)}
}

func (ctx *Context) ThrowError(err error) Value { return ctx.Throw(ctx.Error(err)) }

func (ctx *Context) ThrowSyntaxError(format string, args ...interface{}) Value {
	cause := fmt.Sprintf(format, args...)
	causePtr := C.CString(cause)
	defer C.free(unsafe.Pointer(causePtr))
	return Value{ctx: ctx, ref: C.ThrowSyntaxError(ctx.ref, causePtr)}
}

func (ctx *Context) ThrowTypeError(format string, args ...interface{}) Value {
	cause := fmt.Sprintf(format, args...)
	causePtr := C.CString(cause)
	defer C.free(unsafe.Pointer(causePtr))
	return Value{ctx: ctx, ref: C.ThrowTypeError(ctx.ref, causePtr)}
}

func (ctx *Context) ThrowReferenceError(format string, args ...interface{}) Value {
	cause := fmt.Sprintf(format, args...)
	causePtr := C.CString(cause)
	defer C.free(unsafe.Pointer(causePtr))
	return Value{ctx: ctx, ref: C.ThrowReferenceError(ctx.ref, causePtr)}
}

func (ctx *Context) ThrowRangeError(format string, args ...interface{}) Value {
	cause := fmt.Sprintf(format, args...)
	causePtr := C.CString(cause)
	defer C.free(unsafe.Pointer(causePtr))
	return Value{ctx: ctx, ref: C.ThrowRangeError(ctx.ref, causePtr)}
}

func (ctx *Context) ThrowInternalError(format string, args ...interface{}) Value {
	cause := fmt.Sprintf(format, args...)
	causePtr := C.CString(cause)
	defer C.free(unsafe.Pointer(causePtr))
	return Value{ctx: ctx, ref: C.ThrowInternalError(ctx.ref, causePtr)}
}

func (ctx *Context) Exception() error {
	val := Value{ctx: ctx, ref: C.JS_GetException(ctx.ref)}

	defer val.Free()
	return val.Error()
}

func (ctx *Context) Object() Value {
	return Value{ctx: ctx, ref: C.JS_NewObject(ctx.ref)}
}

func (ctx *Context) Array() Value {
	return Value{ctx: ctx, ref: C.JS_NewArray(ctx.ref)}
}

func (ctx *Context) InitStdModule() {
	stdPtr := C.CString("std")
	defer C.free(unsafe.Pointer(stdPtr))
	C.js_init_module_std(ctx.ref, stdPtr)
}

func (ctx *Context) InitOsModule() {
	osPtr := C.CString("os")
	defer C.free(unsafe.Pointer(osPtr))
	C.js_init_module_os(ctx.ref, osPtr)
}

func (ctx *Context) StdHelper() {
	C.js_std_add_helpers(ctx.ref, -1, nil)
}

func (ctx *Context) StdDumpError() {
	C.js_std_dump_error(ctx.ref)
}

type Atom struct {
	ctx *Context
	ref C.JSAtom
}

func (a Atom) Free() { C.JS_FreeAtom(a.ctx.ref, a.ref) }

func (a Atom) String() string {
	ptr := C.JS_AtomToCString(a.ctx.ref, a.ref)
	defer C.JS_FreeCString(a.ctx.ref, ptr)
	return C.GoString(ptr)
}

func (a Atom) Value() Value {
	return Value{ctx: a.ctx, ref: C.JS_AtomToValue(a.ctx.ref, a.ref)}
}

type Value struct {
	ctx *Context
	ref C.JSValue
}

func (v Value) Free() { C.JS_FreeValue(v.ctx.ref, v.ref) }

func (v Value) Context() *Context { return v.ctx }

func (v Value) Bool() bool { return C.JS_ToBool(v.ctx.ref, v.ref) == 1 }

func (v Value) String() string {
	ptr := C.JS_ToCString(v.ctx.ref, v.ref)
	defer C.JS_FreeCString(v.ctx.ref, ptr)
	return C.GoString(ptr)
}

func (v Value) Int64() int64 {
	val := C.int64_t(0)
	C.JS_ToInt64(v.ctx.ref, &val, v.ref)
	return int64(val)
}

func (v Value) Int32() int32 {
	val := C.int32_t(0)
	C.JS_ToInt32(v.ctx.ref, &val, v.ref)
	return int32(val)
}

func (v Value) Uint32() uint32 {
	val := C.uint32_t(0)
	C.JS_ToUint32(v.ctx.ref, &val, v.ref)
	return uint32(val)
}

func (v Value) Float64() float64 {
	val := C.double(0)
	C.JS_ToFloat64(v.ctx.ref, &val, v.ref)
	return float64(val)
}

func (v Value) BigInt() *big.Int {
	if !v.IsBigInt() {
		return nil
	}
	val, ok := new(big.Int).SetString(v.String(), 10)
	if !ok {
		return nil
	}
	return val
}

func (v Value) BigFloat() *big.Float {
	if !v.IsBigDecimal() && !v.IsBigFloat() {
		return nil
	}
	val, ok := new(big.Float).SetString(v.String())
	if !ok {
		return nil
	}
	return val
}

func (v Value) Get(name string) Value {
	namePtr := C.CString(name)
	defer C.free(unsafe.Pointer(namePtr))
	return Value{ctx: v.ctx, ref: C.JS_GetPropertyStr(v.ctx.ref, v.ref, namePtr)}
}

func (v Value) GetByAtom(atom Atom) Value {
	return Value{ctx: v.ctx, ref: C.JS_GetProperty(v.ctx.ref, v.ref, atom.ref)}
}

func (v Value) GetByUint32(idx uint32) Value {
	return Value{ctx: v.ctx, ref: C.JS_GetPropertyUint32(v.ctx.ref, v.ref, C.uint32_t(idx))}
}

func (v Value) SetByAtom(atom Atom, val Value) {
	C.JS_SetProperty(v.ctx.ref, v.ref, atom.ref, val.ref)
}

func (v Value) SetByInt64(idx int64, val Value) {
	C.JS_SetPropertyInt64(v.ctx.ref, v.ref, C.int64_t(idx), val.ref)
}

func (v Value) SetByUint32(idx uint32, val Value) {
	C.JS_SetPropertyUint32(v.ctx.ref, v.ref, C.uint32_t(idx), val.ref)
}

func (v Value) Len() int64 { return v.Get("length").Int64() }

func (v Value) Set(name string, val Value) {
	namePtr := C.CString(name)
	defer C.free(unsafe.Pointer(namePtr))
	C.JS_SetPropertyStr(v.ctx.ref, v.ref, namePtr, val.ref)
}

func (v Value) SetFunction(name string, fn Function) {
	fnJ := v.ctx.Function(fn)
	v.Set(name, fnJ)
}

type Error struct {
	Cause      string
	Message    string
	FileName   string
	LineNumber string
	Stack      string
}

func (err Error) String() string {
	return fmt.Sprintf("cause:%s,message:%s,filename:%s,linenumber:%s,stack:%s", err.Cause, err.Message, err.FileName, err.LineNumber, err.Stack)
}

func (err Error) Error() string { return err.Cause }

func (v Value) Error() error {
	if !v.IsError() {
		return nil
	}

	cause := v.String()

	message := v.Get("message")
	defer message.Free()

	filename := v.Get("fileName")
	defer filename.Free()

	linenumber := v.Get("lineNumber")
	defer linenumber.Free()

	stack := v.Get("stack")
	defer stack.Free()

	if stack.IsUndefined() {
		return &Error{Cause: cause}
	}
	return &Error{Cause: cause, Message: message.String(), FileName: filename.String(), LineNumber: linenumber.String(), Stack: stack.String()}
}

func (v Value) IsNumber() bool        { return C.JS_IsNumber(v.ref) == 1 }
func (v Value) IsBigInt() bool        { return C.JS_IsBigInt(v.ctx.ref, v.ref) == 1 }
func (v Value) IsBigFloat() bool      { return C.JS_IsBigFloat(v.ref) == 1 }
func (v Value) IsBigDecimal() bool    { return C.JS_IsBigDecimal(v.ref) == 1 }
func (v Value) IsBool() bool          { return C.JS_IsBool(v.ref) == 1 }
func (v Value) IsNull() bool          { return C.JS_IsNull(v.ref) == 1 }
func (v Value) IsUndefined() bool     { return C.JS_IsUndefined(v.ref) == 1 }
func (v Value) IsException() bool     { return C.JS_IsException(v.ref) == 1 }
func (v Value) IsUninitialized() bool { return C.JS_IsUninitialized(v.ref) == 1 }
func (v Value) IsString() bool        { return C.JS_IsString(v.ref) == 1 }
func (v Value) IsSymbol() bool        { return C.JS_IsSymbol(v.ref) == 1 }
func (v Value) IsObject() bool        { return C.JS_IsObject(v.ref) == 1 }
func (v Value) IsArray() bool         { return C.JS_IsArray(v.ctx.ref, v.ref) == 1 }

func (v Value) IsError() bool       { return C.JS_IsError(v.ctx.ref, v.ref) == 1 }
func (v Value) IsFunction() bool    { return C.JS_IsFunction(v.ctx.ref, v.ref) == 1 }
func (v Value) IsConstructor() bool { return C.JS_IsConstructor(v.ctx.ref, v.ref) == 1 }

type PropertyEnum struct {
	IsEnumerable bool
	Atom         Atom
}

func (p PropertyEnum) String() string { return p.Atom.String() }

func (v Value) PropertyNames() ([]PropertyEnum, error) {
	var (
		ptr  *C.JSPropertyEnum
		size C.uint32_t
	)

	result := int(C.JS_GetOwnPropertyNames(v.ctx.ref, &ptr, &size, v.ref, C.int(1<<0|1<<1|1<<2)))
	if result < 0 {
		return nil, errors.New("value does not contain properties")
	}
	defer C.js_free(v.ctx.ref, unsafe.Pointer(ptr))

	entries := (*[1 << unsafe.Sizeof(0)]C.JSPropertyEnum)(unsafe.Pointer(ptr))

	names := make([]PropertyEnum, uint32(size))

	for i := 0; i < len(names); i++ {
		names[i].IsEnumerable = entries[i].is_enumerable == 1

		names[i].Atom = Atom{ctx: v.ctx, ref: entries[i].atom}
		names[i].Atom.Free()
	}

	return names, nil
}

type ModuleInitFn func(ctx *Context, module *Module) int
type Module struct {
	ctx    *Context
	ref    *C.JSModuleDef
	initFn ModuleInitFn
}

func storeModulePtr(v *Module) ObjectId {
	return NewObjectId(v)
}

func restoreModulePtr(id ObjectId) *Module {
	if v, ok := id.Get(); ok {
		if _v, ok := v.(*Module); ok {
			return _v
		}
	}

	return nil
}

//export moduleInitProxy
func moduleInitProxy(ctx *C.JSContext, mod *C.JSModuleDef) C.int {
	meta := C.JS_GetImportMeta(ctx, mod)
	defer C.JS_FreeValue(ctx, meta)
	ptr := C.CString("__goModuleId")
	defer C.free(unsafe.Pointer(ptr))
	vId := C.JS_GetPropertyStr(ctx, meta, ptr)
	id := C.int(0)
	C.JS_ToInt32(ctx, &id, vId)
	module := restoreModulePtr(ObjectId(id))
	return C.int(module.initFn(module.ctx, module))
}

func (ctx *Context) DefineModule(name string, fn ModuleInitFn) *Module {
	ptr := C.CString(name)
	defer C.free(unsafe.Pointer(ptr))
	modRef := C.JS_NewCModule(ctx.ref, ptr, (*C.JSModuleInitFunc)(unsafe.Pointer(C.moduleInitProxy)))
	module := &Module{
		ref:    modRef,
		ctx:    ctx,
		initFn: fn,
	}
	id := storeModulePtr(module)
	meta := C.JS_GetImportMeta(ctx.ref, modRef)
	defer C.JS_FreeValue(ctx.ref, meta)
	ptr2 := C.CString("__goModuleId")
	defer C.free(unsafe.Pointer(ptr2))
	idInt := C.JS_NewInt32(ctx.ref, C.int(id))
	C.JS_SetPropertyStr(ctx.ref, meta, ptr2, idInt)

	return module
}

func (mod *Module) AddFunction(name string, fn Function) {
	ptr := C.CString(name)
	defer C.free(unsafe.Pointer(ptr))
	jFn := mod.ctx.Function(fn)
	C.JS_SetModuleExport(mod.ctx.ref, mod.ref, ptr, jFn.ref)
}

func (mod *Module) AddProperty(name string, v Value) {
	ptr := C.CString(name)
	defer C.free(unsafe.Pointer(ptr))
	C.JS_SetModuleExport(mod.ctx.ref, mod.ref, ptr, v.ref)
}

func (mod *Module) ExportName(name string) {
	ptr := C.CString(name)
	defer C.free(unsafe.Pointer(ptr))
	C.JS_AddModuleExport(mod.ctx.ref, mod.ref, ptr)
}

func (mod *Module) Ref() unsafe.Pointer {
	return unsafe.Pointer(mod.ref)
}

func (ctx *Context) EvaluateFile(name string) int {
	ptr := C.CString(name)
	defer C.free(unsafe.Pointer(ptr))
	return int(C.eval_file(ctx.ref, ptr, -1))
}

type FinalizerFn func(rt *Runtime, val Value)

type Class struct {
	id             ClassId
	ctx            *Context
	defRef         *C.JSClassDef
	proto          *Value
	finalizer      FinalizerFn
	definitionCtor *Value
}

func storeClassPtr(v *Class) ObjectId {
	return NewObjectId(v)
}

func restoreClassPtr(id ObjectId) *Class {
	if v, ok := id.Get(); ok {
		if _v, ok := v.(*Class); ok {
			return _v
		}
	}

	return nil
}

//export finalizeObjectInstance
func finalizeObjectInstance(rt *C.JSRuntime, val C.JSValue) {
	cId := C.ObjectClassID(val)
	id := ClassId(cId)
	cls, ok := id.Get()
	if !ok {
		return
	}
	v := Value{ctx: cls.ctx, ref: val}
	if cls.finalizer == nil {
		return
	}
	cls.finalizer(WrapRuntime(unsafe.Pointer(rt)), v)
}

func newClassId(name string) ClassId {
	id, ok := FindClassId(name)
	if ok {
		return id
	}
	var x C.JSClassID
	C.JS_NewClassID(&x)
	id = ClassId(x)
	AddClassId(name, id)
	return id
}

func (ctx *Context) Runtime() *Runtime {
	r := ctx.RuntimeRef()
	return WrapRuntime(unsafe.Pointer(r))
}
func (ctx *Context) RuntimeRef() *C.JSRuntime {
	return C.JS_GetRuntime(ctx.ref)
}

func (ctx *Context) NewClass(name string, finalizer FinalizerFn) *Class {
	id := newClassId(name)
	cDef := C.NewClassDef(C.CString(name))
	C.JS_NewClass(ctx.RuntimeRef(), C.uint32_t(id), cDef)
	proto := ctx.NewObject()
	cls := &Class{
		id:        id,
		ctx:       ctx,
		defRef:    cDef,
		finalizer: finalizer,
		proto:     &proto,
	}
	AddClassObject(id, cls)
	return cls
}

func (cls *Class) Proto() *Value {
	return cls.proto
}

func (cls *Class) Id() ClassId {
	return cls.id
}

func (cls *Class) ProtoStabilize() {
	C.JS_SetClassProto(cls.ctx.ref, C.uint32_t(cls.id), cls.proto.ref)
}

func (cls *Class) DefConstructor(fn Function) Value {
	classR := cls.ctx.ConstructorFn(fn)
	C.JS_SetConstructor(cls.ctx.ref, classR.ref, cls.proto.ref)
	cls.definitionCtor = &classR
	return *cls.definitionCtor
}

func (cls *Class) Free() {
	cls.proto.Free()
}

func (ctx *Context) NewObject() Value {
	return Value{ctx: ctx, ref: C.JS_NewObject(ctx.ref)}
}

func (cls *Class) NewObject() Value {
	return Value{ctx: cls.ctx, ref: C.JS_NewObjectProtoClass(cls.ctx.ref, cls.proto.ref, C.uint32_t(cls.id))}
}

func (v Value) SetOpaque(data unsafe.Pointer) {
	C.JS_SetOpaque(v.ref, data)
}

func (v Value) GetOpaque(id ClassId) unsafe.Pointer {
	return unsafe.Pointer(C.JS_GetOpaque2(v.ctx.ref, v.ref, C.uint32_t(id)))
}

func (ctx *Context) ConstructorFn(fn Function) Value {
	val := ctx.eval(`(proxy, id) => function() { return proxy.call(this, id, ...arguments); }`)
	if val.IsException() {
		return val
	}
	defer val.Free()

	funcPtr := storeFuncPtr(&funcEntry{ctx: ctx, fn: fn})
	funcPtrVal := ctx.Int64(int64(funcPtr))

	if ctx.ctorProxy == nil {
		ctx.ctorProxy = &Value{
			ctx: ctx,
			ref: C.JS_NewCFunction2(ctx.ref, (*C.JSCFunction)(unsafe.Pointer(C.proxy)), nil, C.int(0), C.JS_CFUNC_generic, C.int(0)),
		}
	}

	args := []C.JSValue{ctx.ctorProxy.ref, funcPtrVal.ref}

	return Value{ctx: ctx, ref: C.JS_Call(ctx.ref, val.ref, ctx.Null().ref, C.int(len(args)), &args[0])}
}

func (ctx *Context) NewArrayBuf(b []byte) Value {
	// TODO: Provide Custom Finalizer Fn
	buf := (*C.uint8_t)(C.CBytes(b))
	jsBuf := C.JS_NewArrayBuffer(ctx.ref, buf, C.size_t(len(b)), (*C.JSFreeArrayBufferDataFunc)(unsafe.Pointer(C.freeArrayBuffer)), nil, C.int(0))
	return Value{
		ctx: ctx,
		ref: jsBuf,
	}
}

func (ctx *Context) NewArrayBufCopy(b []byte) Value {
	ptr := C.CBytes(b)
	buf := (*C.uint8_t)(ptr)
	defer C.free(ptr)
	jsBuf := C.JS_NewArrayBufferCopy(ctx.ref, buf, C.size_t(len(b)))
	return Value{
		ctx: ctx,
		ref: jsBuf,
	}
}

//export freeArrayBuffer
func freeArrayBuffer(rt *C.JSRuntime, opaque unsafe.Pointer, ptr unsafe.Pointer) {
	C.free(ptr)
}
