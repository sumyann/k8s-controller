#!/bin/bash
curl -X POST http://localhost:30098/cache/my-new-key -d 'hello world'
curl -X PUT http://localhost:30098/cache/my-new-key -d 'hello world'
curl http://localhost:30098/cache/my-new-key