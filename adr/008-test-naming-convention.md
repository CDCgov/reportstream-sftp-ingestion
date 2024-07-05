# 8. Test Naming Convention

Date: 2024-07-02

## Decision

For consistency, we'll name tests like
`Test_TheMethodBeingTested_TheCoveredScenario_TheExpectedResult`, e.g.
`Test_Unzip_FileIsPasswordProtected_UnzipsSuccessfully`. For a baseline scenario test,
it's okay to do something like `Test_TheMethodBeingTested_TheExpectedResult`
e.g. `Test_getToken_ReturnsAccessToken`.

## Status

Accepted.
