# GfSpBaseApp

GfSpBaseApp implements the GfSp Framework, it is the entry of program.
GfSpBaseApp starts Grpc server and specify the use of Grpc for communication 
between modules, because the SP is a set of microservices, different modules 
can be deployed in different processes in any combination.

GfSpBaseApp only implements specific processes, which are the standard part 
of GfSp Framework, and the unnecessary parts are customized by modular.

Non-standard processes can register Grpc service to GfSpBaseApp by calling
ServerForRegister, see [retriever](../modular/retriever/retriever.go) as 
example. 

GfSpBaseApp also implements the all [Core Infrastructure Interface](../core/README.md).
GfSpBaseApp will call default [Modular](../modular) that implements 
[Core Special Modular](../core/README.md) to complete the SP ask approval,
upload object payload data, download object and so on workflow.