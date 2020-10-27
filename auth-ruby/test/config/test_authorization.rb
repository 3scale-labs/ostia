# frozen_string_literal: true

require 'minitest/autorun'
require_relative '../test_helper'
require 'json'

describe Config::Authorization do
  let(:authorization_config) { YAML.load(file_fixture('config.yml').read).dig('localhost:8000', 'authorization') }

  describe 'opa' do
    let(:described_class) { Config::Authorization::OPA }
    let(:input) { JSON.parse(file_fixture('opa_input.json').read)['input'] }
    let(:config) { authorization_config.first['opa'] }
    let(:context) do
      OpenStruct.new(to_h: {
        request: input.slice('attributes'),
        identity: { 'test' => input.dig('context', 'identity') },
        metadata: input.dig('context', 'metadata')
      })
    end

    subject do
      described_class.new(config)
    end

    it 'wraps the response' do
      result = subject.call(context)
      expect(result).must_be_instance_of(described_class::Response)
      refute(result.authorized?) # user does not have permission
    end
  end
end
