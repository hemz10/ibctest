Feature: Create chain
    In order to demo IBC
    As a tester
    I need to create a pair of IBC enabled chains

    @smoke
    Scenario: Sending token transfer over IBC
    Given couple of IBC chains running
    And user wallet is funded 
    When we send IBC transfer and relay packets
    Then funds should be transferred and amount should be debitted from account

    Scenario: sending amount more than balance
    Given couple of IBC chains running
    And user wallet is funded
    When we send amount greater than balance
    Then transaction should fail


    Scenario: Sending negative amount for transer
    Given couple of IBC chains running
    And user wallet is funded
    When we send negative amount for transer
    Then transaction should fail




