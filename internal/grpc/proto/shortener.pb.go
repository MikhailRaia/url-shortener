package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

type URLShortenRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	URL string `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *URLShortenRequest) Reset() {
	*x = URLShortenRequest{}
}

func (x *URLShortenRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*URLShortenRequest) ProtoMessage() {}

func (x *URLShortenRequest) ProtoReflect() protoreflect.Message {
	mi := &fileProtoShortenerProtoMsgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *URLShortenRequest) Descriptor() ([]byte, []int) {
	return fileProtoShortenerProtoRawDescGZIP(), []int{0}
}

type URLShortenResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result string `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *URLShortenResponse) Reset() {
	*x = URLShortenResponse{}
}

func (x *URLShortenResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*URLShortenResponse) ProtoMessage() {}

func (x *URLShortenResponse) ProtoReflect() protoreflect.Message {
	mi := &fileProtoShortenerProtoMsgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *URLShortenResponse) Descriptor() ([]byte, []int) {
	return fileProtoShortenerProtoRawDescGZIP(), []int{1}
}

type URLExpandRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *URLExpandRequest) Reset() {
	*x = URLExpandRequest{}
}

func (x *URLExpandRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*URLExpandRequest) ProtoMessage() {}

func (x *URLExpandRequest) ProtoReflect() protoreflect.Message {
	mi := &fileProtoShortenerProtoMsgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *URLExpandRequest) Descriptor() ([]byte, []int) {
	return fileProtoShortenerProtoRawDescGZIP(), []int{2}
}

type URLExpandResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result string `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *URLExpandResponse) Reset() {
	*x = URLExpandResponse{}
}

func (x *URLExpandResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*URLExpandResponse) ProtoMessage() {}

func (x *URLExpandResponse) ProtoReflect() protoreflect.Message {
	mi := &fileProtoShortenerProtoMsgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *URLExpandResponse) Descriptor() ([]byte, []int) {
	return fileProtoShortenerProtoRawDescGZIP(), []int{3}
}

type URLData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ShortURL    string `protobuf:"bytes,1,opt,name=short_url,json=shortUrl,proto3" json:"short_url,omitempty"`
	OriginalURL string `protobuf:"bytes,2,opt,name=original_url,json=originalUrl,proto3" json:"original_url,omitempty"`
}

func (x *URLData) Reset() {
	*x = URLData{}
}

func (x *URLData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*URLData) ProtoMessage() {}

func (x *URLData) ProtoReflect() protoreflect.Message {
	mi := &fileProtoShortenerProtoMsgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *URLData) Descriptor() ([]byte, []int) {
	return fileProtoShortenerProtoRawDescGZIP(), []int{4}
}

type UserURLsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	URL []*URLData `protobuf:"bytes,1,rep,name=url,proto3" json:"url,omitempty"`
}

func (x *UserURLsResponse) Reset() {
	*x = UserURLsResponse{}
}

func (x *UserURLsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserURLsResponse) ProtoMessage() {}

func (x *UserURLsResponse) ProtoReflect() protoreflect.Message {
	mi := &fileProtoShortenerProtoMsgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *UserURLsResponse) Descriptor() ([]byte, []int) {
	return fileProtoShortenerProtoRawDescGZIP(), []int{5}
}

var fileProtoShortenerProtoMsgTypes = make([]protoimpl.MessageInfo, 6)

func fileProtoShortenerProtoRawDescGZIP() []byte {
	return nil
}

var FileProtoShortenerProto protoreflect.FileDescriptor
