cmake_minimum_required(VERSION 3.30)
project(libroutey C)

set(CMAKE_C_STANDARD 11)

add_library(libroutey STATIC library.c)
