use crate::counter::Counter;
use crate::limit::Limit;
use crate::storage::{InMemoryStorage, Storage};
use std::collections::{HashMap, HashSet};
use std::error::Error;
use std::fmt;
use std::fmt::{Display, Formatter};

mod counter;
pub mod limit;
mod storage;

#[derive(Debug)]
pub struct MissingNamespaceErr {
    msg: String,
}

impl MissingNamespaceErr {
    fn new() -> MissingNamespaceErr {
        MissingNamespaceErr {
            msg: "Missing namespace".to_string(),
        }
    }
}

impl Display for MissingNamespaceErr {
    fn fmt(&self, f: &mut Formatter) -> fmt::Result {
        write!(f, "{}", self.msg)
    }
}

impl Error for MissingNamespaceErr {
    fn description(&self) -> &str {
        &self.msg
    }
}

pub struct RateLimiter {
    storage: InMemoryStorage, // TODO: use trait here
}

impl RateLimiter {
    pub fn new() -> RateLimiter {
        RateLimiter {
            storage: InMemoryStorage::new(),
        }
    }

    pub fn add_limit(&mut self, limit: Limit) {
        self.storage.add_limit(limit);
    }

    pub fn get_limits(&self, namespace: &str) -> HashSet<Limit> {
        self.storage.get_limits(namespace)
    }

    pub fn is_rate_limited(
        &self,
        values: &HashMap<String, String>,
    ) -> Result<bool, MissingNamespaceErr> {
        // TODO: hardcoded delta

        match values.get("namespace") {
            Some(namespace) => Ok(self
                .get_limits(namespace)
                .iter()
                .filter(|lim| lim.applies(values))
                .map(|lim| Counter::new(lim.clone(), values.clone()))
                .any(|counter| !self.storage.is_within_limits(&counter, 1))),
            None => Err(MissingNamespaceErr::new()),
        }
    }

    pub fn update_counters(
        &mut self,
        values: &HashMap<String, String>,
    ) -> Result<(), MissingNamespaceErr> {
        match values.get("namespace") {
            // TODO: hardcoded delta
            Some(namespace) => Ok(self
                .get_limits(namespace)
                .iter()
                .filter(|lim| lim.applies(values))
                .map(|lim| Counter::new(lim.clone(), values.clone()))
                .for_each(|counter| self.storage.update_counter(&counter, 1))),
            None => Err(MissingNamespaceErr::new()),
        }
    }
}
