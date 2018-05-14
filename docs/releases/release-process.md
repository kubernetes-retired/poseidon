# Motivation
Due to the fact that Poseidon/Firmament scheduler project has to deal with two distinct Git repositories 
(one for Poseidon repo under K8S incubation org & another one for Firmament repo under CamSas), 
there is a need for coordinated release process in order for these two repos to work together in lockstep. 
This document captures details for the overall release plan for Poseidon/Firmament scheduling environment.

# Goals
* Come up with a software release plan for Poseidon and Firmament such that K8s community might find it easy to use it.

# User Stories
### Story 1
I want to use Poseidon and Firmament as an alternate pluggable scheduler for K8s. Which release version should I use? How do I get the release binaries? What version of Firmament works with what version of Poseidon?

# Proposed Solution
* For a major new feature development within Firmament repo, a new Firmament release package will be created. 
  At the same time, a new Poseidon release package will be created that corresponds to the newly created Firmament 
  release package (Poseidon deployment manifest in the new release package points to the new Firmament release artifacts).
  A new Poseidon release package will be created in this case irrespective of the fact that there are any corresponding 
  changes to Poseidon repo or not.
  
* For a major new feature development within Poseidon repo, a new Poseidon release package will be created.
  In case there are corresponding changes to the Firmament codebase, the new Poseidon release package will point to the 
  new Firmament release package. Otherwise, new Poseidon release package will point to the previous Firmament release 
  package. Essentially, there is no need to cut a Firmament release if there are no changes to Firmament repo.
  
* Poseidon release notes for each release will document what version of Poseidon works with corresponding Firmament 
  release version. 
