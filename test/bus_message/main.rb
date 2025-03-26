require 'message_bus'
require 'redis'

# Configure MessageBus to use your custom Redis server
MessageBus.configure(
  backend: :redis,
  redis_config: {
    url: "redis://127.0.0.1:6479/0"
  },
)

# Subscribe to a channel
Thread.new do
  MessageBus.subscribe('/test_channel') do |msg|
    puts "[Subscriber] Received message: #{msg.data}"
  end
end

# Give the subscriber some time to start
sleep 1

# Publish a message
MessageBus.publish('/test_channel', 'Hello from MessageBus!')

# Keep the script running to receive messages
sleep 5
