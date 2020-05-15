use crate::counter::Counter;
use crate::limit::Limit;
use crate::storage::{Storage, StorageErr};
use std::collections::{HashMap, HashSet};
use std::time::Duration;
use ttl_cache::TtlCache;

pub struct InMemoryStorage {
    limits_for_namespace: HashMap<String, HashSet<Limit>>,
    counters: TtlCache<Counter, i64>,
}

impl Storage for InMemoryStorage {
    fn add_limit(&mut self, limit: Limit) -> Result<(), StorageErr> {
        let namespace = limit.namespace().to_string();

        match self.limits_for_namespace.get_mut(&namespace) {
            Some(value) => {
                value.insert(limit);
            }
            None => {
                let mut limits = HashSet::new();
                limits.insert(limit);
                self.limits_for_namespace.insert(namespace, limits);
            }
        }

        Ok(())
    }

    fn get_limits(&self, namespace: &str) -> Result<HashSet<Limit>, StorageErr> {
        let limits = match self.limits_for_namespace.get(namespace) {
            Some(limits) => limits.clone(),
            None => HashSet::new(),
        };

        Ok(limits)
    }

    fn is_within_limits(&self, counter: &Counter, delta: i64) -> Result<bool, StorageErr> {
        let within_limits = match self.counters.get(counter) {
            Some(value) => *value - delta >= 0,
            None => true,
        };

        Ok(within_limits)
    }

    fn update_counter(&mut self, counter: &Counter, delta: i64) -> Result<(), StorageErr> {
        match self.counters.get_mut(counter) {
            Some(value) => {
                *value -= delta;
            }
            None => {
                self.counters.insert(
                    counter.clone(),
                    counter.max_value() - 1,
                    Duration::from_secs(counter.seconds()),
                );
            }
        };

        Ok(())
    }
}

impl InMemoryStorage {
    pub fn new() -> InMemoryStorage {
        InMemoryStorage {
            limits_for_namespace: HashMap::new(),
            counters: TtlCache::new(1000), // TODO: adjust capacity using ENV
        }
    }
}
