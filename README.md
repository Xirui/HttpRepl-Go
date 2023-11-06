# HttpRepl-Go
Golang implementation of [dotnet HttpRepl](https://github.com/dotnet/HttpRepl).

There is a great value of having a lightweight, cross-platform command-line tool for making HTTP requests to test OpenAPIs and view their results. The original implementation requires .NET Core runtime which could be a limiting factor in many scenarios.  A simple executable binary would argubly be a better choice, especially for IoT and edge computing applications.

Thus, efforts are put into making a golang implementation of HttpRepl. The aim, however, is not to have 100% feature-compatible version.