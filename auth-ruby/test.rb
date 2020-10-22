require 'minitest/autorun'
require_relative 'auth'
require 'json'

describe V2AuthorizationService do
  let(:service) { V2AuthorizationService.new }

  describe 'request' do
    let(:body) do
      {
        'attributes' => {
          'source' => { 'address' => { 'socketAddress' => { 'protocol' => 'TCP', 'address' => '172.18.0.1', 'portValue' => 47_428, 'resolverName' => '', 'ipv4Compat' => false } }, 'service' => '', 'labels' => {}, 'principal' => '', 'certificate' => '' },
          'destination' => {},
          'request' => {
            'time' => '2020-10-22T11:55:24.500288000Z', 'http' => {
              'id' => '8911960713991399467', 'method' => 'GET', 'headers' => {
                ':path' => '/foo', ':method' => 'GET', ':authority' => 'localhost:8000',
                'x-request-id' => '7a9bc5dd-1672-43ca-a99f-e209b8c725ca', 'user-agent' => 'curl/7.70.0',
                'x-forwarded-proto' => 'http', 'authorization' => 'Basic Zm9vOmJhcg==', 'accept' => '*/*'
              }, 'path' => '/foo', 'host' => 'localhost:8000', 'scheme' => '', 'query' => '', 'fragment' => '', 'size' => '0', 'protocol' => 'HTTP/1.1', 'body' => ''
            }
          }, 'contextExtensions' => {}, 'metadataContext' => {
            'filterMetadata' => {}
          }
        }
      }
    end

    let(:request) do
      Envoy::Service::Auth::V2::CheckRequest.decode_json body.to_json
    end

    subject do
      service.check(request, GRPC::ActiveCall.allocate)
    end

    it 'must respond positively' do
      expect(subject).must_be_instance_of(Envoy::Service::Auth::V2::CheckResponse)
    end
  end
end
