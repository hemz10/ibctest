Feature: Create Connection
    In order to connect 2 chains via IBC
    As a Relayer
    After client is created then Connection should be established

    Scenario: Create Connection
    Given couple of IBC chains running
    And relay creates a path
    And client should be created on both chains
    When relay creates a connection
    Then connection should be established

