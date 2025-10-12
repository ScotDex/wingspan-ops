# Short Circuit - Future Features Roadmap

## 🟢 Easy (Quick Wins & Visual Polish)
- [ ] Display more wormhole info (Class, Size, Mass) in route results.
- [ ] Generate an "FC, please help!" copy-paste string for the calculated route.
- [ ] Add basic wormhole filters to the form (e.g., "Ignore EOL", "Ignore Critical Mass").

---

## 🟡 Medium (Requires EVE Login)
- [ ] Integrate the full "Log in with EVE" OAuth2 flow.
- [ ] Add a "Use my current location" button for the route planner's start system.
- [ ] Add a "Set in-game destination" button to the route results.

---

## 🔴 Hard (Core Logic Changes)
- [ ] Implement Security Prioritization sliders (this requires adding "weights" to the pathfinding algorithm).
- [ ] Implement a system "Avoidance List".
- [ ] Integrate with the Tripwire API to pull chain maps.