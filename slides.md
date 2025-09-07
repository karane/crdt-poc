% Conflict-free Replicated Data Types (CRDTs)
% Karane Vieira
% 2025-07-10

# Conflict-free Replicated Data Types (CRDTs)

---

## The Problem

- In distributed systems, data is replicated across nodes.
- Nodes may be offline, slow, or partitioned.
- Traditional approaches:
  - **Locks / Consensus (e.g., Paxos, Raft)** → consistent but expensive.
  - **Eventual Consistency** → fast but resolving conflicts is hard.

**Question:** How can we replicate data safely *without central coordination*?

---

## The Idea of CRDTs

- **CRDTs are data structures designed for distributed systems.**
- They guarantee **strong eventual consistency**.
- **Key property:** Updates can be applied in any order, on any replica, and all replicas **converge** to the same state.
- Achieved via **mathematical merge rules** that are:
  - Commutative
  - Associative
  - Idempotent

---

## Two Flavors of CRDTs

1. **State-based (Convergent / CvRDTs):**
   - Each replica occasionally sends its full state.
   - **Merge** 

2. **Operation-based (Commutative / CmRDTs):**
   - Each replica sends **operations to other replicas**.
   - Operations are commutative and idempotent.

---

## Examples of CRDTs – Counters

- **G-Counter (Grow-only Counter)**
  - Only increments
  - Use case: Counting likes, views, votes

- **PN-Counter (Positive/Negative Counter)**
  - Increments & decrements
  - Use case: Inventory systems, balances

---

## Examples of CRDTs – Sets

- **G-Set (Grow-only Set)**
  - Only adds elements
  - Use case: Membership lists, feature flags

- **OR-Set (Observed-Removed Set)**
  - Adds & removes elements with unique tags
  - Use case: Collaborative tagging, shared todo lists

---

## Examples of CRDTs – Registers

- **Last-Writer-Wins Register (LWW)**
  - Resolves conflicts using timestamp
  - Use case: User profiles, configuration values

---

## Examples of CRDTs – Sequences

\small

Used for collaborative text editting  

- **RGA (Replicated Growable Array)**  
  - Text as a linked list  
  - Linearizes insertions and deletions using unique IDs and timestamps  
  - Simple and good for small and medium docs

- **Logoot**
  - Uses global ids for positions
  - Doesn´t require tombstones for deletions
  - Good for large scale docs

---

## Examples of CRDTs – Sequences - More

\small

- **WOOT (WithOut Operational Transform)**
  - Doubly linked list of characters
  - Deletions with milestones
  - Handles complex concurrent editing scenarios

- **LSEQ (List SEQuence)**
  - Optimizes identifier allocation to keep IDs short
  - Reduces overhead compared to Logoot in large documents
  - Provides logarithmic growth of identifiers even with many insertions
  - Good for large docs

---

## Why CRDTs Matter

- **No locks, no central coordination.**
- Tolerates **network partitions**.
- Perfect for:
  - **Real-time collaboration** (Google Docs, Figma, Notion)
  - **Geo-replicated databases** (Riak, Redis CRDTs, AntidoteDB)
  - **Offline-first apps** (messaging apps, note-taking apps)
- Each CRDT type is suited for a specific use case:
  - Counters → metrics, likes, inventory  
  - Sets → membership, shared collections  
  - Registers → configs, profiles  
  - Sequences → text, code, documents

---

## Libraries Using CRDTs

- **Collaboration / Real-time Editors**  
  -- **Automerge** – JavaScript CRDT library for collaborative applications  
  -- **Yjs** – CRDT library for collaborative text, spreadsheets, and graphics  

---

## Limitations

- Not all data types fit easily into CRDTs.
- Can have **higher memory overhead** (metadata, tombstones).
- Sometimes more complex than traditional approaches.

---

## And Now !?!?

Show me the code.