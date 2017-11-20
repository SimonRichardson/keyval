# Key/Value Store

## Command keyval

 - [Getting started](#getting-started)
 - [Introduction](#introduction)
 - [Store](#store)
 - [Network](#network)
 - [Tests](#tests)
 - [Improvements](#improvements)

### Getting started

The Key/Val store CLI expects a few prerequisites to be installed for it to be
able to build. A valid `go` installation (1.8+) and a valid `$GOPATH`.

The quickest way to get up and running is to execute the following:

```
make install dist/keyval
./dist/keyval store
```

The go to the following url [localhost:8080](http://localhost:8080/store/?key=abc)

### Introduction

The keyval code base is split up into a number of parts:

  1. Store: This contains the internal memory key value store
  2. HTTP: This represents a http rest end point to the store
  3. TCP: A tcp server for the store
  4. UDP: And finally the udp server for the store

### Store

The store implementation uses the internal map implementation under the hood,
doesn't really do anything special, but requires that all keys can be
represented via a string. This is so that we can use the key with in the url
of the end points.

There also exists a bucket version of the map, which is left in as an exercise
to show that without little effort we can get a parallelised read/write
implementation, yet still align to the same store interface.

### Network

The networking part of the code base can be mainly thought of as two parts. The
http version and the tcp/udp versions. Both version use the same underlying
store so you can query from all protocols and get the same data back.

The http version just uses the HTTP methods GET/PUT/DELETE to get/set/delete
respectively. I didn't opt for a router here as the end points and methods are
easily implemented. Alternatively something like gorilla mux or similar could
easily be used to fill this space.

For both tcp/udp versions, they both use golang gob for encoding and decoding
data. This offers some really quick wins without much effort for speed and
reliability whilst still being really abstracted from the types required in the
store directly.

### Tests

The tests with in the project use various types of testing, to show more of a
variety than anything else. So we use gomock to create mocks for expectations
of what the store can do. We also have quick checking (fuzzing) and finally
benchmark testing for the store itself.

### Improvements

  1. More testing, esp. around edge cases and errors, but because of the short
  time frame around this, some things where cut.
  2. Better UDP buffer allocation. The buffer allocation for the UDP server is
  really greedy, because of the nature of the values that can be put in the
  store. We could use some other methods to allow better streaming of data to
  prevent resource drainage.
