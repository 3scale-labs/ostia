require "application_system_test_case"

class ClientsTest < ApplicationSystemTestCase
  setup do
    @client = clients(:one)
  end

  test "visiting the index" do
    visit clients_url
    assert_selector "h1", text: "Clients"
  end

  test "creating a Client" do
    visit clients_url
    click_on "New Client"

    fill_in "Client", with: @client.client_id
    fill_in "Client secret", with: @client.client_secret
    fill_in "Registration access token", with: @client.registration_access_token
    fill_in "Registration uri", with: @client.registration_uri
    fill_in "User", with: @client.user_id
    click_on "Create Client"

    assert_text "Client was successfully created"
    click_on "Back"
  end

  test "updating a Client" do
    visit clients_url
    click_on "Edit", match: :first

    fill_in "Client", with: @client.client_id
    fill_in "Client secret", with: @client.client_secret
    fill_in "Registration access token", with: @client.registration_access_token
    fill_in "Registration uri", with: @client.registration_uri
    fill_in "User", with: @client.user_id
    click_on "Update Client"

    assert_text "Client was successfully updated"
    click_on "Back"
  end

  test "destroying a Client" do
    visit clients_url
    page.accept_confirm do
      click_on "Destroy", match: :first
    end

    assert_text "Client was successfully destroyed"
  end
end
