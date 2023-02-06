Feature: Create Client
    In order to connect 2 chains via IBC
    As a Relayer
    First step is to create clients on both chains

    Scenario: Create Client
    Given couple of IBC chains running
    When relay creates a path
    Then client should be created on both chains

    Scenario: Query Client ID
    Given couple of IBC chains running
    And relay creates a path
    And client should be created on both chains
    When we query client
    Then Client ID should be returned

    Scenario: Query Client Status
    Given couple of IBC chains running
    And relay creates a path
    And client should be created on both chains
    When we query client
    Then Client status should be returned

   
    Scenario: Query latest height from client state
    Given couple of IBC chains running
    And relay creates a path
    And client should be created on both chains
    When we query client
    Then latest height from client state should be returned

    Scenario: Create Client with not existing path
    Given couple of IBC chains running
    And relay creates a path
    When we create relay with not existing path
    Then client should not be created

    Scenario: Query for all clients
    Given couple of IBC chains running
    And relay creates a path
    And client should be created on both chains
    When we query client
    Then all number of clients should be returned

    Scenario: Update Client
    Given couple of IBC chains running
    And relay creates a path
    And client should be created on both chains
    And we query client
    When we update client
    And we query client
    Then client should be updated

    @smoke
    Scenario: Query for relay account balance
    Given couple of IBC chains running
    And relay creates a path
    And client should be created on both chains
    When we query using relay
    Then relay account balance should be returned

    