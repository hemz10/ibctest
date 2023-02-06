Feature: Create Channel
    In order to connect 2 chains via IBC
    As a Relayer
    After client and connection is created then Channel should be established

    Scenario: Create Connection
    Given couple of IBC chains running
    And relay creates a path
    And client should be created on both chains
    And relay creates a connection
    When channel is created
    Then channel should be established
