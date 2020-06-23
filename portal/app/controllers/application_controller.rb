class ApplicationController < ActionController::Base
  before_action :current_user

  helper_method :current_user

  protected

  def current_identity
    @current_identity ||= GlobalID::Locator.locate(session[:identity_gid])
  end

  def current_user
    @current_user ||= current_identity&.user
  rescue ActiveRecord::RecordNotFound
    reset_session
  end
end
