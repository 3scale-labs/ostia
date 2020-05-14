#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StatusAnnotation {
    /// The entity is work-in-progress and subject to breaking changes.
    #[prost(bool, tag = "1")]
    pub work_in_progress: bool,
    /// The entity belongs to a package with the given version status.
    #[prost(enumeration = "PackageVersionStatus", tag = "2")]
    pub package_version_status: i32,
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum PackageVersionStatus {
    /// Unknown package version status.
    Unknown = 0,
    /// This version of the package is frozen.
    Frozen = 1,
    /// This version of the package is the active development version.
    Active = 2,
    /// This version of the package is the candidate for the next major version. It
    /// is typically machine generated from the active development version.
    NextMajorVersionCandidate = 3,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct VersioningAnnotation {
    /// Track the previous message type. E.g. this message might be
    /// udpa.foo.v3alpha.Foo and it was previously udpa.bar.v2.Bar. This
    /// information is consumed by UDPA via proto descriptors.
    #[prost(string, tag = "1")]
    pub previous_message_type: std::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MigrateAnnotation {
    /// Rename the message/enum/enum value in next version.
    #[prost(string, tag = "1")]
    pub rename: std::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FieldMigrateAnnotation {
    /// Rename the field in next version.
    #[prost(string, tag = "1")]
    pub rename: std::string::String,
    /// Add the field to a named oneof in next version. If this already exists, the
    /// field will join its siblings under the oneof, otherwise a new oneof will be
    /// created with the given name.
    #[prost(string, tag = "2")]
    pub oneof_promotion: std::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FileMigrateAnnotation {
    /// Move all types in the file to another package, this implies changing proto
    /// file path.
    #[prost(string, tag = "2")]
    pub move_to_package: std::string::String,
}
