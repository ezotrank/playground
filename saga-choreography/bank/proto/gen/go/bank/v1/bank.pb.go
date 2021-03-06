// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        (unknown)
// source: bank/v1/bank.proto

package bankv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Account_Status int32

const (
	Account_STATUS_UNSPECIFIED Account_Status = 0
	Account_STATUS_NEW         Account_Status = 1
	Account_STATUS_REGISTERED  Account_Status = 2
	Account_STATUS_REJECTED    Account_Status = 3
)

// Enum value maps for Account_Status.
var (
	Account_Status_name = map[int32]string{
		0: "STATUS_UNSPECIFIED",
		1: "STATUS_NEW",
		2: "STATUS_REGISTERED",
		3: "STATUS_REJECTED",
	}
	Account_Status_value = map[string]int32{
		"STATUS_UNSPECIFIED": 0,
		"STATUS_NEW":         1,
		"STATUS_REGISTERED":  2,
		"STATUS_REJECTED":    3,
	}
)

func (x Account_Status) Enum() *Account_Status {
	p := new(Account_Status)
	*p = x
	return p
}

func (x Account_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Account_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_bank_v1_bank_proto_enumTypes[0].Descriptor()
}

func (Account_Status) Type() protoreflect.EnumType {
	return &file_bank_v1_bank_proto_enumTypes[0]
}

func (x Account_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Account_Status.Descriptor instead.
func (Account_Status) EnumDescriptor() ([]byte, []int) {
	return file_bank_v1_bank_proto_rawDescGZIP(), []int{0, 0}
}

type Transaction_Status int32

const (
	Transaction_STATUS_UNSPECIFIED Transaction_Status = 0
	Transaction_STATUS_SUCCEED     Transaction_Status = 1
	Transaction_STATUS_REJECTED    Transaction_Status = 2
)

// Enum value maps for Transaction_Status.
var (
	Transaction_Status_name = map[int32]string{
		0: "STATUS_UNSPECIFIED",
		1: "STATUS_SUCCEED",
		2: "STATUS_REJECTED",
	}
	Transaction_Status_value = map[string]int32{
		"STATUS_UNSPECIFIED": 0,
		"STATUS_SUCCEED":     1,
		"STATUS_REJECTED":    2,
	}
)

func (x Transaction_Status) Enum() *Transaction_Status {
	p := new(Transaction_Status)
	*p = x
	return p
}

func (x Transaction_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Transaction_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_bank_v1_bank_proto_enumTypes[1].Descriptor()
}

func (Transaction_Status) Type() protoreflect.EnumType {
	return &file_bank_v1_bank_proto_enumTypes[1]
}

func (x Transaction_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Transaction_Status.Descriptor instead.
func (Transaction_Status) EnumDescriptor() ([]byte, []int) {
	return file_bank_v1_bank_proto_rawDescGZIP(), []int{1, 0}
}

type Account struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AccountId    string         `protobuf:"bytes,1,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
	UserId       string         `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Status       Account_Status `protobuf:"varint,3,opt,name=status,proto3,enum=bank.v1.Account_Status" json:"status,omitempty"`
	RejectReason string         `protobuf:"bytes,4,opt,name=reject_reason,json=rejectReason,proto3" json:"reject_reason,omitempty"`
}

func (x *Account) Reset() {
	*x = Account{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bank_v1_bank_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Account) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Account) ProtoMessage() {}

func (x *Account) ProtoReflect() protoreflect.Message {
	mi := &file_bank_v1_bank_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Account.ProtoReflect.Descriptor instead.
func (*Account) Descriptor() ([]byte, []int) {
	return file_bank_v1_bank_proto_rawDescGZIP(), []int{0}
}

func (x *Account) GetAccountId() string {
	if x != nil {
		return x.AccountId
	}
	return ""
}

func (x *Account) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *Account) GetStatus() Account_Status {
	if x != nil {
		return x.Status
	}
	return Account_STATUS_UNSPECIFIED
}

func (x *Account) GetRejectReason() string {
	if x != nil {
		return x.RejectReason
	}
	return ""
}

type Transaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionId string             `protobuf:"bytes,1,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	AccountId     string             `protobuf:"bytes,2,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
	Amount        int64              `protobuf:"varint,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Status        Transaction_Status `protobuf:"varint,4,opt,name=status,proto3,enum=bank.v1.Transaction_Status" json:"status,omitempty"`
	RejectReason  string             `protobuf:"bytes,5,opt,name=reject_reason,json=rejectReason,proto3" json:"reject_reason,omitempty"`
}

func (x *Transaction) Reset() {
	*x = Transaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bank_v1_bank_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Transaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction) ProtoMessage() {}

func (x *Transaction) ProtoReflect() protoreflect.Message {
	mi := &file_bank_v1_bank_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction.ProtoReflect.Descriptor instead.
func (*Transaction) Descriptor() ([]byte, []int) {
	return file_bank_v1_bank_proto_rawDescGZIP(), []int{1}
}

func (x *Transaction) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

func (x *Transaction) GetAccountId() string {
	if x != nil {
		return x.AccountId
	}
	return ""
}

func (x *Transaction) GetAmount() int64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *Transaction) GetStatus() Transaction_Status {
	if x != nil {
		return x.Status
	}
	return Transaction_STATUS_UNSPECIFIED
}

func (x *Transaction) GetRejectReason() string {
	if x != nil {
		return x.RejectReason
	}
	return ""
}

var File_bank_v1_bank_proto protoreflect.FileDescriptor

var file_bank_v1_bank_proto_rawDesc = []byte{
	0x0a, 0x12, 0x62, 0x61, 0x6e, 0x6b, 0x2f, 0x76, 0x31, 0x2f, 0x62, 0x61, 0x6e, 0x6b, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x61, 0x6e, 0x6b, 0x2e, 0x76, 0x31, 0x22, 0xf5, 0x01,
	0x0a, 0x07, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72,
	0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49,
	0x64, 0x12, 0x2f, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x17, 0x2e, 0x62, 0x61, 0x6e, 0x6b, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x72, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x72, 0x65, 0x61,
	0x73, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65, 0x6a, 0x65, 0x63,
	0x74, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x22, 0x5c, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x16, 0x0a, 0x12, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x55, 0x4e, 0x53, 0x50,
	0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x53, 0x54, 0x41,
	0x54, 0x55, 0x53, 0x5f, 0x4e, 0x45, 0x57, 0x10, 0x01, 0x12, 0x15, 0x0a, 0x11, 0x53, 0x54, 0x41,
	0x54, 0x55, 0x53, 0x5f, 0x52, 0x45, 0x47, 0x49, 0x53, 0x54, 0x45, 0x52, 0x45, 0x44, 0x10, 0x02,
	0x12, 0x13, 0x0a, 0x0f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x52, 0x45, 0x4a, 0x45, 0x43,
	0x54, 0x45, 0x44, 0x10, 0x03, 0x22, 0x90, 0x02, 0x0a, 0x0b, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x0e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a,
	0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x61,
	0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x61, 0x6d, 0x6f,
	0x75, 0x6e, 0x74, 0x12, 0x33, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x62, 0x61, 0x6e, 0x6b, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x72, 0x65, 0x6a, 0x65,
	0x63, 0x74, 0x5f, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0c, 0x72, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x22, 0x49, 0x0a,
	0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x16, 0x0a, 0x12, 0x53, 0x54, 0x41, 0x54, 0x55,
	0x53, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12,
	0x12, 0x0a, 0x0e, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x53, 0x55, 0x43, 0x43, 0x45, 0x45,
	0x44, 0x10, 0x01, 0x12, 0x13, 0x0a, 0x0f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x52, 0x45,
	0x4a, 0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x02, 0x42, 0xa8, 0x01, 0x0a, 0x0b, 0x63, 0x6f, 0x6d,
	0x2e, 0x62, 0x61, 0x6e, 0x6b, 0x2e, 0x76, 0x31, 0x42, 0x09, 0x42, 0x61, 0x6e, 0x6b, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x51, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x65, 0x7a, 0x6f, 0x74, 0x72, 0x61, 0x6e, 0x6b, 0x2f, 0x70, 0x6c, 0x61, 0x79, 0x67,
	0x72, 0x6f, 0x75, 0x6e, 0x64, 0x2f, 0x73, 0x61, 0x67, 0x61, 0x2d, 0x63, 0x68, 0x6f, 0x72, 0x65,
	0x6f, 0x67, 0x72, 0x61, 0x70, 0x68, 0x79, 0x2f, 0x62, 0x61, 0x6e, 0x6b, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x6f, 0x2f, 0x62, 0x61, 0x6e, 0x6b, 0x2f, 0x76,
	0x31, 0x3b, 0x62, 0x61, 0x6e, 0x6b, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x42, 0x58, 0x58, 0xaa, 0x02,
	0x07, 0x42, 0x61, 0x6e, 0x6b, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x07, 0x42, 0x61, 0x6e, 0x6b, 0x5c,
	0x56, 0x31, 0xe2, 0x02, 0x13, 0x42, 0x61, 0x6e, 0x6b, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x08, 0x42, 0x61, 0x6e, 0x6b, 0x3a,
	0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_bank_v1_bank_proto_rawDescOnce sync.Once
	file_bank_v1_bank_proto_rawDescData = file_bank_v1_bank_proto_rawDesc
)

func file_bank_v1_bank_proto_rawDescGZIP() []byte {
	file_bank_v1_bank_proto_rawDescOnce.Do(func() {
		file_bank_v1_bank_proto_rawDescData = protoimpl.X.CompressGZIP(file_bank_v1_bank_proto_rawDescData)
	})
	return file_bank_v1_bank_proto_rawDescData
}

var file_bank_v1_bank_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_bank_v1_bank_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_bank_v1_bank_proto_goTypes = []interface{}{
	(Account_Status)(0),     // 0: bank.v1.Account.Status
	(Transaction_Status)(0), // 1: bank.v1.Transaction.Status
	(*Account)(nil),         // 2: bank.v1.Account
	(*Transaction)(nil),     // 3: bank.v1.Transaction
}
var file_bank_v1_bank_proto_depIdxs = []int32{
	0, // 0: bank.v1.Account.status:type_name -> bank.v1.Account.Status
	1, // 1: bank.v1.Transaction.status:type_name -> bank.v1.Transaction.Status
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_bank_v1_bank_proto_init() }
func file_bank_v1_bank_proto_init() {
	if File_bank_v1_bank_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_bank_v1_bank_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Account); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_bank_v1_bank_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Transaction); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_bank_v1_bank_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_bank_v1_bank_proto_goTypes,
		DependencyIndexes: file_bank_v1_bank_proto_depIdxs,
		EnumInfos:         file_bank_v1_bank_proto_enumTypes,
		MessageInfos:      file_bank_v1_bank_proto_msgTypes,
	}.Build()
	File_bank_v1_bank_proto = out.File
	file_bank_v1_bank_proto_rawDesc = nil
	file_bank_v1_bank_proto_goTypes = nil
	file_bank_v1_bank_proto_depIdxs = nil
}
