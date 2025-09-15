package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type ProtoServiceDesc struct {
	ServiceName string
	Methods     []ProtoMethodDesc
}

type ProtoMethodDesc struct {
	MethodName     string
	InputType      string
	OutputType     string
	IsClientStream bool
	IsServerStream bool
}

// ReadProtoFileAndExtractServices reads a proto file and extracts all gRPC service descriptors
func ReadProtoFileAndExtractServices(protoFilePath string) ([]ProtoServiceDesc, error) {
	// Get the directory containing the proto file
	protoDir := filepath.Dir(protoFilePath)
	protoFileName := filepath.Base(protoFilePath)

	// Create a temporary file for the descriptor
	tempFile, err := os.CreateTemp("", "proto_descriptor_*.pb")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Build the protoc command with proper import paths
	// Include common protobuf import paths
	cmd := exec.Command("protoc",
		"--proto_path="+protoDir,
		"--proto_path=/usr/include",       // Common system include path
		"--proto_path=/usr/local/include", // Another common include path
		"--include_imports",
		"--include_source_info",
		"--descriptor_set_out="+tempFile.Name(),
		protoFileName)

	// Run protoc
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("protoc failed: %w, stderr: %s", err, stderr.String())
	}

	// Read the generated descriptor file
	descriptorData, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read descriptor file: %w", err)
	}

	// Parse the descriptor file
	fileDesc, err := parseDescriptorFile(descriptorData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse descriptor file: %w", err)
	}

	// Extract service descriptors
	return extractServiceDescriptors(fileDesc), nil
}

func parseDescriptorFile(descriptorData []byte) (protoreflect.FileDescriptor, error) {
	// Parse the descriptor set
	descriptorSet := &descriptorpb.FileDescriptorSet{}
	err := proto.Unmarshal(descriptorData, descriptorSet)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal descriptor set: %w", err)
	}

	if len(descriptorSet.File) == 0 {
		return nil, fmt.Errorf("no files in descriptor set")
	}

	// Create file descriptors for all files in the set
	fileDescs := make([]protoreflect.FileDescriptor, len(descriptorSet.File))
	for i, fd := range descriptorSet.File {
		fileDesc, err := protodesc.NewFile(fd, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create file descriptor for %s: %w", fd.GetName(), err)
		}
		fileDescs[i] = fileDesc
	}

	// Return the first file descriptor (main proto file)
	return fileDescs[0], nil
}

func extractServiceDescriptors(fileDesc protoreflect.FileDescriptor) []ProtoServiceDesc {
	var serviceDescs []ProtoServiceDesc
	services := fileDesc.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		serviceDesc := ProtoServiceDesc{
			ServiceName: string(service.Name()),
			Methods:     extractMethods(service),
		}
		serviceDescs = append(serviceDescs, serviceDesc)
	}
	return serviceDescs
}

func extractMethods(service protoreflect.ServiceDescriptor) []ProtoMethodDesc {
	var methods []ProtoMethodDesc

	serviceMethods := service.Methods()
	for i := 0; i < serviceMethods.Len(); i++ {
		method := serviceMethods.Get(i)
		methodDesc := ProtoMethodDesc{
			MethodName:     string(method.Name()),
			InputType:      string(method.Input().FullName()),
			OutputType:     string(method.Output().FullName()),
			IsClientStream: method.IsStreamingClient(),
			IsServerStream: method.IsStreamingServer(),
		}
		methods = append(methods, methodDesc)
	}

	return methods
}
