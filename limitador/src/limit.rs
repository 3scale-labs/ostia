use serde::Deserialize;
use std::collections::{HashMap, HashSet};
use std::hash::{Hash, Hasher};

#[derive(Eq, PartialEq, Debug, Clone, Deserialize)]
pub struct Limit {
    namespace: String,
    max_value: i64,
    seconds: u64,
    conditions: HashSet<String>,
    variables: HashSet<String>,
}

impl Limit {
    pub fn new<S: Into<String>>(
        namespace: S,
        max_value: i64,
        seconds: u64,
        conditions: HashSet<String>,
        variables: HashSet<String>,
    ) -> Limit {
        Limit {
            namespace: namespace.into(),
            max_value,
            seconds,
            conditions,
            variables,
        }
    }

    pub fn namespace(&self) -> &str {
        &self.namespace
    }

    pub fn max_value(&self) -> i64 {
        self.max_value
    }

    pub fn seconds(&self) -> u64 {
        self.seconds
    }

    pub fn applies(&self, values: &HashMap<String, String>) -> bool {
        self.conditions
            .iter()
            .all(|cond| Self::condition_applies(&cond, values))
    }

    fn condition_applies(condition: &str, values: &HashMap<String, String>) -> bool {
        // TODO: for now assume that all the conditions have this format:
        // "left_operand == right_operand"

        let split: Vec<&str> = condition.split(" == ").collect();
        let left_operand = split[0];
        let right_operand = split[1];

        &values[left_operand] == right_operand
    }
}

impl Hash for Limit {
    fn hash<H: Hasher>(&self, state: &mut H) {
        self.namespace.hash(state);
        self.max_value.hash(state);
        self.seconds.hash(state);

        self.conditions
            .iter()
            .map(|c| c.to_string())
            .collect::<Vec<String>>()
            .sort()
            .hash(state);

        self.variables
            .iter()
            .map(|c| c.to_string())
            .collect::<Vec<String>>()
            .sort()
            .hash(state);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn limit_applies() {
        let mut conditions: HashSet<String> = HashSet::new();
        conditions.insert("x == 5".into());

        let mut variables: HashSet<String> = HashSet::new();
        variables.insert("y".into());

        let limit = Limit::new("test_namespace", 10, 60, conditions, variables);

        let mut values: HashMap<String, String> = HashMap::new();
        values.insert("x".into(), "5".into());

        assert!(limit.applies(&values))
    }

    #[test]
    fn limit_does_not_apply() {
        let mut conditions: HashSet<String> = HashSet::new();
        conditions.insert("x == 5".into());

        let mut variables: HashSet<String> = HashSet::new();
        variables.insert("y".into());

        let limit = Limit::new("test_namespace", 10, 60, conditions, variables);

        let mut values: HashMap<String, String> = HashMap::new();
        values.insert("x".into(), "1".into());

        assert_eq!(false, limit.applies(&values))
    }

    // TODO: test limits with multiple conditions
}
