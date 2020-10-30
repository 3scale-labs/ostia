require 'minitest/autorun'
require_relative 'test_helper'
require 'json'
require 'config'

describe V2AuthorizationService do
  let(:config) { Config.new(file_fixture('config.yml')) }
  let(:service) { V2AuthorizationService.new(config) }

  describe 'request' do
    let(:body) { JSON.parse(file_fixture('request.json').read) }
    let(:request) { Envoy::Service::Auth::V2::CheckRequest.decode_json body.to_json }

    subject do
      service.check(request, GRPC::ActiveCall.allocate)
    end

    it 'must respond positively' do
      expect(subject).must_be_instance_of(Envoy::Service::Auth::V2::CheckResponse)
      expect(subject.ok_response).must_be_instance_of(Envoy::Service::Auth::V2::OkHttpResponse)
    end
  end
end
