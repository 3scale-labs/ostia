use crate::counter::Counter;
use crate::errors::LimitadorError;
use crate::limit::Limit;
use crate::storage::in_memory::InMemoryStorage;
use crate::storage::Storage;
use std::collections::{HashMap, HashSet};

mod counter;
pub mod errors;
pub mod limit;
mod storage;

pub struct RateLimiter {
    storage: Box<dyn Storage>,
}

impl RateLimiter {
    pub fn new() -> RateLimiter {
        RateLimiter {
            storage: Box::new(InMemoryStorage::new()),
        }
    }

    pub fn add_limit(&mut self, limit: Limit) -> Result<(), LimitadorError> {
        self.storage.add_limit(limit).map_err(|err| err.into())
    }

    pub fn get_limits(&self, namespace: &str) -> Result<HashSet<Limit>, LimitadorError> {
        self.storage.get_limits(namespace).map_err(|err| err.into())
    }

    pub fn is_rate_limited(
        &self,
        values: &HashMap<String, String>,
    ) -> Result<bool, LimitadorError> {
        // TODO: hardcoded delta

        match values.get("namespace") {
            Some(namespace) => {
                let counters = self.counters_that_apply(namespace, values)?;

                for counter in counters {
                    match self.storage.is_within_limits(&counter, 1) {
                        Ok(within_limits) => {
                            if !within_limits {
                                return Ok(true);
                            }
                        }
                        Err(e) => return Err(e.into()),
                    }
                }

                Ok(false)
            }
            None => Err(LimitadorError::MissingNamespace),
        }
    }

    pub fn update_counters(
        &mut self,
        values: &HashMap<String, String>,
    ) -> Result<(), LimitadorError> {
        match values.get("namespace") {
            // TODO: hardcoded delta
            Some(namespace) => {
                let counters = self.counters_that_apply(namespace, values)?;

                counters
                    .iter()
                    .try_for_each(|counter| self.storage.update_counter(&counter, 1))
                    .map_err(|err| err.into())
            }
            None => Err(LimitadorError::MissingNamespace),
        }
    }

    // TODO: return iterator and avoid collect() call
    fn counters_that_apply(
        &self,
        namespace: &str,
        values: &HashMap<String, String>,
    ) -> Result<Vec<Counter>, LimitadorError> {
        let limits = self.get_limits(namespace)?;

        let counters = limits
            .iter()
            .filter(|lim| lim.applies(values))
            .map(|lim| Counter::new(lim.clone(), values.clone()))
            .collect();

        Ok(counters)
    }
}

impl Default for RateLimiter {
    fn default() -> Self {
        Self::new()
    }
}
