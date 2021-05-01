import hashlib
import os
import warnings

# From Third Party
import click


# From This Project
import config
import ucrypt
from udb import UDB


class DefaultGroup(click.Group):
    """Invokes a subcommand marked with `default=True` if any subcommand not
    chosen.
    :param default_if_no_args: resolves to the default command if no arguments
                               passed.
    """

    def __init__(self, *args, **kwargs):
        # To resolve as the default command.
        if not kwargs.get('ignore_unknown_options', True):
            raise ValueError('Default group accepts unknown options')
        self.ignore_unknown_options = True
        self.default_cmd_name = kwargs.pop('default', None)
        self.default_if_no_args = kwargs.pop('default_if_no_args', False)
        super(DefaultGroup, self).__init__(*args, **kwargs)

    def set_default_command(self, command):
        """Sets a command function as the default command."""
        cmd_name = command.name
        self.add_command(command)
        self.default_cmd_name = cmd_name

    def parse_args(self, ctx, args):
        if not args and self.default_if_no_args:
            args.insert(0, self.default_cmd_name)
        return super(DefaultGroup, self).parse_args(ctx, args)

    def get_command(self, ctx, cmd_name):
        if cmd_name not in self.commands:
            # No command name matched.
            ctx.arg0 = cmd_name
            cmd_name = self.default_cmd_name
        return super(DefaultGroup, self).get_command(ctx, cmd_name)

    def resolve_command(self, ctx, args):
        base = super(DefaultGroup, self)
        cmd_name, cmd, args = base.resolve_command(ctx, args)
        if hasattr(ctx, 'arg0'):
            args.insert(0, ctx.arg0)
            cmd_name = cmd.name
        return cmd_name, cmd, args

    def format_commands(self, ctx, formatter):
        formatter = DefaultCommandFormatter(self, formatter, mark='*')
        return super(DefaultGroup, self).format_commands(ctx, formatter)

    def command(self, *args, **kwargs):
        default = kwargs.pop('default', False)
        decorator = super(DefaultGroup, self).command(*args, **kwargs)
        if not default:
            return decorator
        warnings.warn('Use default param of DefaultGroup or '
                      'set_default_command() instead', DeprecationWarning)

        def _decorator(f):
            cmd = decorator(f)
            self.set_default_command(cmd)
            return cmd

        return _decorator


class DefaultCommandFormatter(object):
    """Wraps a formatter to mark a default command."""

    def __init__(self, group, formatter, mark='*'):
        self.group = group
        self.formatter = formatter
        self.mark = mark

    def __getattr__(self, attr):
        return getattr(self.formatter, attr)

    def write_dl(self, rows, *args, **kwargs):
        rows_ = []
        for cmd_name, help in rows:
            if cmd_name == self.group.default_cmd_name:
                rows_.insert(0, (cmd_name + self.mark, help))
            else:
                rows_.append((cmd_name, help))
        return self.formatter.write_dl(rows_, *args, **kwargs)

def sha2_id(s):
    if (not s) or (not type(s) is str):
        return None
    return hashlib.sha256(s.encode("UTF-8")).hexdigest()


def used_name_lookup(cur_name, db_path="", latest=True):
    if not db_path:
        db_path = os.path.join(config.gParamDict["record_path"], "rd.db")
    _cur_id = sha2_id(cur_name)
    db = UDB(db_path)
    df = db.checkout_rd(_cur_id)
    db.close()
    if "curCrypt" not in df.columns:
        return None
    rlt = list(df["curCrypt"])
    if len(rlt) == 0:
        return None
    if latest:
        return ucrypt.b64_str_decrypt(rlt[0], cur_name)
    else:
        return [
            ucrypt.b64_str_decrypt(elm, cur_name)
            for elm in list(dict.fromkeys(rlt))
        ]


def log_to_db(cur_name, new_name, db_path=""):
    if not db_path:
        db_path = os.path.join(config.gParamDict["record_path"], "rd.db")
    _cur_id = sha2_id(cur_name)
    _new_id = sha2_id(new_name)
    _cur_crypt = ucrypt.encrypt_b64_str(cur_name, new_name)
    db = UDB(db_path)
    db.insert_rd(_new_id, _cur_id, _cur_crypt)
    db.close()
