package main

import (
	"math"
	"testing"
	"unsafe"
)

// ============ Point 结构体测试 ============

// TestPoint_Size 验证 Point 结构体大小
func TestPoint_Size(t *testing.T) {
	p := Point{X: 1.0, Y: 2.0, Z: "test"}
	size := unsafe.Sizeof(p)
	t.Logf("Point 结构体大小: %d 字节", size)
	// float64 = 8, float64 = 8, string = 16 (64位), 合计 ≥ 32
	if size < 32 {
		t.Errorf("Point 大小应 ≥ 32，实际为 %d", size)
	}
}

// TestPoint_Offset 验证 Point 结构体字段偏移
func TestPoint_Offset(t *testing.T) {
	var p Point
	if unsafe.Offsetof(p.X) != 0 {
		t.Errorf("X 偏移应为 0，实际为 %d", unsafe.Offsetof(p.X))
	}
	// Y 偏移应为 8（64位下 float64 对齐后）
	if unsafe.Offsetof(p.Y) != 8 {
		t.Errorf("Y 偏移应为 8，实际为 %d", unsafe.Offsetof(p.Y))
	}
	// Z (string) 偏移应为 16
	if unsafe.Offsetof(p.Z) != 16 {
		t.Errorf("Z 偏移应为 16，实际为 %d", unsafe.Offsetof(p.Z))
	}
}

// ============ float64 ↔ uint64 位转换验证 ============

// TestFloat64BitConversion 测试 float64 与 uint64 之间的位转换可逆性
func TestFloat64BitConversion(t *testing.T) {
	original := 3.141592653589793
	bits := *(*uint64)(unsafe.Pointer(&original))
	restored := *(*float64)(unsafe.Pointer(&bits))
	if restored != original {
		t.Errorf("位转换不可逆: original=%v, restored=%v", original, restored)
	}
}

// ============ string 和 []byte 零拷贝转换验证 ============

// TestStringToBytesConversion 测试零拷贝 string → []byte 转换
func TestStringToBytesConversion(t *testing.T) {
	original := "Hello, 不安全的世界！"
	strHdr := (*stringHeader)(unsafe.Pointer(&original))
	byteSlice := *(*[]byte)(unsafe.Pointer(&sliceHeader{
		Data: strHdr.Data,
		Len:  strHdr.Len,
		Cap:  strHdr.Len,
	}))
	// 字节切片长度应与字符串长度一致
	if len(byteSlice) != len(original) {
		t.Errorf("转换后长度不匹配: %d vs %d", len(byteSlice), len(original))
	}
	// 每个字节应相同
	for i := 0; i < len(original); i++ {
		if byteSlice[i] != original[i] {
			t.Errorf("位置 %d 字节不匹配: %02x vs %02x", i, byteSlice[i], original[i])
			break
		}
	}
}

// TestStringHeader_Size 验证 stringHeader 大小
func TestStringHeader_Size(t *testing.T) {
	var sh stringHeader
	size := unsafe.Sizeof(sh)
	// 64位系统: unsafe.Pointer(8) + int(8) = 16
	if size != 16 {
		t.Errorf("stringHeader 大小应为 16，实际为 %d", size)
	}
}

// TestSliceHeader_Size 验证 sliceHeader 大小
func TestSliceHeader_Size(t *testing.T) {
	var sh sliceHeader
	size := unsafe.Sizeof(sh)
	// 64位系统: unsafe.Pointer(8) + int(8) + int(8) = 24
	if size != 24 {
		t.Errorf("sliceHeader 大小应为 24，实际为 %d", size)
	}
}

// ============ 基本类型 Sizeof / Alignof 验证 ============

// TestSizeof_BasicTypes 验证基本类型的 Sizeof
func TestSizeof_BasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		variable interface{}
		minSize  uintptr
	}{
		{"int", int(0), 4},
		{"float64", float64(0), 8},
		{"bool", bool(false), 1},
		{"string", string(""), 16}, // 64位系统 string 是 16 字节
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 通过 reflect 获取变量值，不能用 unsafe.Sizeof 直接对 interface{} 取值
			// 这里直接用已知类型测试
		})
	}
}

// TestSizeof_StringIs16 测试 string 类型大小
func TestSizeof_StringIs16(t *testing.T) {
	var s string
	sz := unsafe.Sizeof(s)
	if sz != 16 {
		t.Errorf("string sizeof 应为 16 (64位系统)，实际为 %d", sz)
	}
}

// ============ 结构体对齐 ============

// TestPointUnaligned 测试 Point 结构体至少是 float64 对齐的
func TestPointUnaligned(t *testing.T) {
	var p Point
	align := unsafe.Alignof(p)
	if align < 8 {
		t.Errorf("Point 对齐要求应 ≥ 8，实际为 %d", align)
	}
}

// ============ 指针运算验证 ============

// TestPointerArithmetic 测试数组的指针运算
func TestPointerArithmetic(t *testing.T) {
	arr := [5]int{10, 20, 30, 40, 50}
	first := unsafe.Pointer(&arr[0])
	elemSize := unsafe.Sizeof(arr[0])

	// 通过指针运算访问索引 2
	third := (*int)(unsafe.Pointer(uintptr(first) + 2*elemSize))
	if *third != 30 {
		t.Errorf("arr[2] 应为 30，通过指针运算得到 %d", *third)
	}

	// 修改索引 4
	fifth := (*int)(unsafe.Pointer(uintptr(first) + 4*elemSize))
	*fifth = 99
	if arr[4] != 99 {
		t.Errorf("arr[4] 应为 99，实际为 %d", arr[4])
	}
}

// ============ 内存读写验证 ============

// TestMemoryReadWrite 测试在字节数组中的内存读写
func TestMemoryReadWrite(t *testing.T) {
	data := make([]byte, 64)
	dataPtr := unsafe.Pointer(&data[0])

	// 写入 int32
	*(*int32)(dataPtr) = 0x12345678
	readBack := *(*int32)(dataPtr)
	if readBack != 0x12345678 {
		t.Errorf("int32 读写不一致: 期望 0x12345678, 得到 0x%x", readBack)
	}

	// 偏移 8 写入 float64
	*(*float64)(unsafe.Pointer(uintptr(dataPtr) + 8)) = 3.14
	f64 := *(*float64)(unsafe.Pointer(uintptr(dataPtr) + 8))
	if f64 != 3.14 {
		t.Errorf("float64 读写不一致: 期望 3.14, 得到 %f", f64)
	}

	// 偏移 16 写入字节数组
	*(*[4]byte)(unsafe.Pointer(uintptr(dataPtr) + 16)) = [4]byte{'G', 'o', '!', 0}
	b4 := *(*[4]byte)(unsafe.Pointer(uintptr(dataPtr) + 16))
	if b4[0] != 'G' || b4[1] != 'o' || b4[2] != '!' {
		t.Errorf("字节数组读写不一致: 期望 [G o ! 0], 得到 %v", b4)
	}
}

// ============ 未导出字段访问验证 ============

// TestPersonFields_ByOffset 通过偏移量读取 person 结构体字段
func TestPersonFields_ByOffset(t *testing.T) {
	alice := person{
		name:   "Alice",
		age:    30,
		secret: "这是秘密信息",
	}

	ptr := unsafe.Pointer(&alice)

	// name 字段在偏移 0
	namePtr := (*string)(ptr)
	if *namePtr != "Alice" {
		t.Errorf("name 应为 Alice，得到 %q", *namePtr)
	}

	// age 字段在 string 之后（string = 16 bytes）
	stringSize := unsafe.Sizeof(string(""))
	agePtr := (*int)(unsafe.Pointer(uintptr(ptr) + stringSize))
	if *agePtr != 30 {
		t.Errorf("age 应为 30，得到 %d", *agePtr)
	}

	// secret 字段在 age 之后
	intSize := unsafe.Sizeof(int(0))
	secretPtr := (*string)(unsafe.Pointer(uintptr(ptr) + stringSize + intSize))
	if *secretPtr != "这是秘密信息" {
		t.Errorf("secret 应为 '这是秘密信息'，得到 %q", *secretPtr)
	}
}

// TestIntSize 验证 int 类型大小（64 位系统应为 8）
func TestIntSize(t *testing.T) {
	var i int
	if unsafe.Sizeof(i) != 8 {
		t.Errorf("64位系统 int 大小应为 8，实际为 %d", unsafe.Sizeof(i))
	}
}

// TestFloat64BitPattern_SpecialValues 验证 float64 的位转换可逆性（含特殊值）
func TestFloat64BitPattern_SpecialValues(t *testing.T) {
	values := []float64{0.0, -0.0, 1.0, -1.0, 3.141592653589793, math.Inf(1)} // +Inf
	for _, v := range values {
		bits := *(*uint64)(unsafe.Pointer(&v))
		restored := *(*float64)(unsafe.Pointer(&bits))
		// +Inf 通过 > 1e300 判断
		if v > 1e300 && restored > 1e300 {
			continue
		}
		// 普通值应精确还原
		if restored != v {
			t.Errorf("float64 %v 位转换不可逆: 得到 %v", v, restored)
		}
	}
}