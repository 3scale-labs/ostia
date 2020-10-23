require 'pathname'

root = Pathname(__dir__).expand_path.dirname

$LOAD_PATH << root.join('grpc')
$LOAD_PATH << root.join('src')

require 'auth'
