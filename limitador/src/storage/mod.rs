use crate::counter::Counter;
use crate::limit::Limit;
use std::collections::HashSet;

pub mod in_memory;

pub trait Storage: Sync + Send {
    fn add_limit(&mut self, limit: Limit);
    fn get_limits(&self, namespace: &str) -> HashSet<Limit>;
    fn is_within_limits(&self, counter: &Counter, delta: i64) -> bool;
    fn update_counter(&mut self, counter: &Counter, delta: i64);
}
