// generate random cryptographic hex

require("crypto").randomBytes(256, (err, buffer) => {
   if (err) throw err;
   console.log(buffer.toString("hex"));
});
