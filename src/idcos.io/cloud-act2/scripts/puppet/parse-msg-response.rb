require "base64"
require "json"

message = ARGF.read


message =  Marshal.load(Base64.decode64(message))
message[:body] = Marshal.load(message[:body])
puts message.to_json
