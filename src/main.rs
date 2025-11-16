use std::path::PathBuf;
use std::{fs, process};

use clap::Parser;

/// Simple CLI that accepts a single input path (file or directory).
#[derive(Debug, Parser)]
#[command(
    name = "path-checker",
    version,
    about = "Detects whether the provided input path is a file or a directory.",
    author
)]
struct Args {
    /// Input path to a file or directory
    input: PathBuf,
}

fn main() {
    let args = Args::parse();

    // Use symlink_metadata so symlinks are recognized and not automatically followed.
    let metadata = match fs::symlink_metadata(&args.input) {
        Ok(m) => m,
        Err(err) => {
            eprintln!("Error: cannot access '{}': {}", args.input.display(), err);
            process::exit(2);
        }
    };

    if metadata.is_dir() {
        println!("'{}' is a directory.", args.input.display());
    } else if metadata.is_file() {
        println!("'{}' is a file.", args.input.display());
    } else {
        println!(
            "'{}' exists but is neither a regular file nor a directory.",
            args.input.display()
        );
    }
}
