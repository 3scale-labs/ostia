use crate::limit::Limit;
use std::collections::HashMap;
use std::hash::{Hash, Hasher};

#[derive(Eq, PartialEq, Clone)]
pub struct Counter {
    limit: Limit,
    set_variables: HashMap<String, String>,
}

impl Counter {
    pub fn new(limit: Limit, set_variables: HashMap<String, String>) -> Counter {
        // TODO: check that all the variables defined in the limit are set.

        Counter {
            limit,
            set_variables,
        }
    }

    pub fn max_value(&self) -> i64 {
        self.limit.max_value()
    }

    pub fn seconds(&self) -> u64 {
        self.limit.seconds()
    }
}

impl Hash for Counter {
    fn hash<H: Hasher>(&self, state: &mut H) {
        self.limit.hash(state);

        self.set_variables
            .iter()
            .map(|(k, v)| k.to_owned() + ":" + v)
            .collect::<Vec<String>>()
            .sort()
            .hash(state);
    }
}
