#!/bin/sh

protoc -I greet/greetpb/ greet/greetpb/greet.proto --go_out=plugins=grpc:greet/greetpb

protoc -I calculator/calculatorpb/ calculator/calculatorpb/calculator.proto --go_out=plugins=grpc:calculator/calculatorpb

protoc -I blog/blogpb/ blog/blogpb/blog.proto --go_out=plugins=grpc:blog/blogpb