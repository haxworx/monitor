#!/usr/bin/env nodejs

// Copyright (c) 2015. Al Poole <netstar@gmail.com>
//
// http://haxlab.org
//
// Proof-of-concept server for dropsy

var http = require('http');
var mkdirp = require('mkdirp');

function respondCode(res, val)
{
    res.writeHead(200, {'Content-Type' : 'text/plain'});
    res.write("status: " + val + "\r\n");
    res.end();
}

http.createServer(function (req, res) {
    var fs = require('fs');

    var username = req.headers['username'];
    var password = req.headers['password'];
    var action = req.headers['action'];
    var filename = req.headers['filename'];
    var directory = req.headers['directory'];

    if (username != "username" || password != "password") {
        respondCode(res, 1);
    } else {
        respondCode(res, 0);
    
    if (action === "ADD") {
        var path = directory + "/" + filename;
        if (!fs.existsSync(directory)) {
            mkdirp(directory, function(err) {
                if (err) console.err(err);
                else console.log("created " + directory);
            });
        }
       
        var outFile = fs.createWriteStream(path);
        req.pipe(outFile);
    } 

    if (action === "DEL") {
        var path = directory + "/" + filename;
        fs.unlink(path, function(err) { 
		if (err) console.err(err);
	});
    }
}
}).listen(12345);

console.log("Copyright (c) 2015. Al Poole <netstar@gmail.com>");
console.log("Drop Server Running!");

